package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MamangRust/microservice-point-of-sale-email/config"
	"github.com/MamangRust/microservice-point-of-sale-email/handler"
	"github.com/MamangRust/microservice-point-of-sale-email/mailer"
	"github.com/MamangRust/microservice-point-of-sale-email/metrics"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/emailretry"
	"github.com/MamangRust/microservice-point-of-sale-pkg/kafka"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	otel_pkg "github.com/MamangRust/microservice-point-of-sale-pkg/otel"
	"github.com/MamangRust/microservice-point-of-sale-pkg/outbox"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	logger, err := logger.NewLogger("email-service", nil)
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}

	if err := dotenv.Viper(); err != nil {
		logger.Fatal("Failed to load .env file", zap.Error(err))
	}

	cfg := config.Config{
		KafkaBrokers: []string{viper.GetString("KAFKA_BROKERS")},
		SMTPServer:   viper.GetString("SMTP_SERVER"),
		SMTPPort:     viper.GetInt("SMTP_PORT"),
		SMTPUser:     viper.GetString("SMTP_USER"),
		SMTPPass:     viper.GetString("SMTP_PASS"),
		MaxRetries:   viper.GetInt("EMAIL_MAX_RETRIES"),
		RetryBackoff: viper.GetDuration("EMAIL_RETRY_BACKOFF"),
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = emailretry.DefaultMaxAttempts
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = emailretry.DefaultBackoff
	}

	telemetry := otel_pkg.NewTelemetry(otel_pkg.Config{
		ServiceName:    "email-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		Insecure:       true,
	})
	if err := telemetry.Init(ctx); err != nil {
		logger.Fatal("Failed to initialize telemetry", zap.Error(err))
	}
	defer func() {
		_ = telemetry.Shutdown(context.Background())
	}()

	metrics.Register()

	m := &mailer.Mailer{
		Server:   cfg.SMTPServer,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Password: cfg.SMTPPass,
	}

	mongoClient, err := database.NewMongoClientForContext(logger, database.EmailDB)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB for consumer inbox", zap.Error(err))
	}
	defer mongoClient.Disconnect(context.Background())

	mongoDB := database.NewMongoDatabase(mongoClient, database.EmailDB)

	inbox, err := outbox.NewMongoInbox(mongoDB)
	if err != nil {
		logger.Fatal("Failed to initialize consumer inbox", zap.Error(err))
	}

	myKafka := kafka.NewKafka(logger, cfg.KafkaBrokers)

	h := handler.NewEmailHandlerWithInbox(m, inbox, "email-service-group", myKafka, cfg.RetryBackoff)

	err = myKafka.StartConsumersWithContext(ctx, []string{
		"email-service-topic-auth-register",
		"email-service-topic-auth-forgot-password",
		"email-service-topic-auth-verify-code-success",
		"email-service-topic-merchant-create",
		"email-service-topic-merchant-update-status",
		"email-service-topic-merchant-document-create",
		"email-service-topic-merchant-document-update-status",
		"email-service-topic-transaction-create",
	}, "email-service-group", h)

	if err != nil {
		log.Fatalf("Error starting consumer: %v", err)
	}

	retryH := handler.NewRetryHandler(m, inbox, "email-service-group", myKafka, cfg.MaxRetries, cfg.RetryBackoff)
	if err := myKafka.StartConsumersWithContext(ctx, []string{emailretry.RetryTopic}, emailretry.RetryGroup, retryH); err != nil {
		log.Fatalf("Error starting retry consumer: %v", err)
	}

	logger.Info("Email service started", zap.String("retry_topic", emailretry.RetryTopic), zap.String("dlq_topic", emailretry.DLQTopic))

	<-ctx.Done()
	logger.Info("Shutting down email service")
	if err := myKafka.Close(); err != nil {
		logger.Error("Failed to close Kafka resources", zap.Error(err))
	}
}
