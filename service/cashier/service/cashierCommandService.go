package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-cashier/cache"
	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	useradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/outbox"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/microservice-point-of-sale-shared/errorhandler"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type cashierCommandDeps struct {
	Cache          mencache.CashierCommandCache
	MerchantQuery  merchantadapter.Repository
	UserQuery      *useradapter.Repository
	CashierCommand repository.CashierCommandRepository
	Outbox         *outbox.OutboxService
	Logger         logger.LoggerInterface
	Observability  observability.TraceLoggerObservability
}

type cashierCommandService struct {
	mencache       mencache.CashierCommandCache
	merchantQuery  merchantadapter.Repository
	userQuery      *useradapter.Repository
	cashierCommand repository.CashierCommandRepository
	outbox         *outbox.OutboxService
	logger         logger.LoggerInterface
	observability  observability.TraceLoggerObservability
}

func NewCashierCommandService(params *cashierCommandDeps) CashierCommandService {
	return &cashierCommandService{
		mencache:       params.Cache,
		merchantQuery:  params.MerchantQuery,
		userQuery:      params.UserQuery,
		cashierCommand: params.CashierCommand,
		outbox:         params.Outbox,
		logger:         params.Logger,
		observability:  params.Observability,
	}
}

// publishStatsEvent marshals a stats domain event into the standard envelope
// (event_id for idempotency + business payload) and enqueues it in the outbox
// for the relay to publish to Kafka. It is best-effort: a failure here is
// logged and reconciled via backfill, never by rolling back the cashier change.
func (s *cashierCommandService) publishStatsEvent(ctx context.Context, cashier *models.Cashier, status string) {
	if s.outbox == nil || cashier == nil {
		return
	}
	const topic = "stats.pos.cashier.event"

	payloadJSON, err := json.Marshal(events.CashierEvent{
		CashierID:  cashier.CashierID,
		MerchantID: cashier.MerchantID,
		Status:     status,
		EventTime:  time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		s.logger.Error("failed to marshal stats event payload", zap.Error(err), zap.String("topic", topic))
		return
	}
	envelope, err := json.Marshal(events.StatsEnvelope{
		EventID: uuid.NewString(),
		Payload: payloadJSON,
	})
	if err != nil {
		s.logger.Error("failed to marshal stats event envelope", zap.Error(err), zap.String("topic", topic))
		return
	}
	if err := s.outbox.Enqueue(ctx, topic, strconv.Itoa(int(cashier.CashierID)), envelope); err != nil {
		s.logger.Error("failed to enqueue stats event", zap.Error(err), zap.String("topic", topic))
	}
}

func (s *cashierCommandService) CreateCashier(ctx context.Context, req *requests.CreateCashierRequest) (*models.Cashier, error) {
	const method = "CreateCashier"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	_, err := s.merchantQuery.FindById(ctx, req.MerchantID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Cashier](
			s.logger,
			merchant_errors.ErrFailedFindMerchantById,
			method,
			span,
			zap.Error(err),
		)
	}

	_, err = s.userQuery.FindById(ctx, req.UserID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Cashier](
			s.logger,
			user_errors.ErrUserNotFoundRes,
			method,
			span,
			zap.Error(err),
		)
	}

	cashier, err := s.cashierCommand.CreateCashier(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Cashier](
			s.logger,
			cashier_errors.ErrFailedCreateCashier,
			method,
			span,
			zap.Error(err),
		)
	}

	// F3 — stats pipeline: publish a stats.pos.cashier.event so stats-writer
	// materializes this cashier into ClickHouse for the cashier statistics.
	s.publishStatsEvent(ctx, cashier, "active")

	logSuccess("Successfully created cashier", zap.Bool("success", true))
	return cashier, nil
}

func (s *cashierCommandService) UpdateCashier(ctx context.Context, req *requests.UpdateCashierRequest) (*models.Cashier, error) {
	const method = "UpdateCashier"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	cashier, err := s.cashierCommand.UpdateCashier(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Cashier](
			s.logger,
			cashier_errors.ErrFailedUpdateCashier,
			method,
			span,
			zap.Error(err),
		)
	}

	span.SetAttributes(attribute.String("cashier.name", cashier.Name))
	s.mencache.DeleteCashierCache(ctx, int(cashier.CashierID))

	s.publishStatsEvent(ctx, cashier, "active")

	logSuccess("Successfully updated cashier", zap.Bool("success", true))
	return cashier, nil
}

func (s *cashierCommandService) TrashedCashier(ctx context.Context, cashierID int) (*models.Cashier, error) {
	const method = "TrashedCashier"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	cashier, err := s.cashierCommand.TrashedCashier(ctx, cashierID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Cashier](
			s.logger,
			cashier_errors.ErrFailedTrashedCashier,
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCashierCache(ctx, cashierID)

	s.publishStatsEvent(ctx, cashier, "trashed")

	logSuccess("Successfully trashed cashier", zap.Bool("success", true))
	return cashier, nil
}

func (s *cashierCommandService) RestoreCashier(ctx context.Context, cashierID int) (*models.Cashier, error) {
	const method = "RestoreCashier"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	cashier, err := s.cashierCommand.RestoreCashier(ctx, cashierID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Cashier](
			s.logger,
			cashier_errors.ErrFailedRestoreCashier,
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCashierCache(ctx, cashierID)

	s.publishStatsEvent(ctx, cashier, "active")

	logSuccess("Successfully restored cashier", zap.Bool("success", true))
	return cashier, nil
}

func (s *cashierCommandService) DeleteCashierPermanent(ctx context.Context, cashierID int) (bool, error) {
	const method = "DeleteCashierPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	success, err := s.cashierCommand.DeleteCashierPermanent(ctx, cashierID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](
			s.logger,
			cashier_errors.ErrFailedDeleteCashierPermanent,
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCashierCache(ctx, cashierID)
	logSuccess("Successfully deleted cashier permanently", zap.Bool("success", success))
	return success, nil
}

func (s *cashierCommandService) RestoreAllCashier(ctx context.Context) (bool, error) {
	const method = "RestoreAllCashier"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	success, err := s.cashierCommand.RestoreAllCashier(ctx)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](
			s.logger,
			cashier_errors.ErrFailedRestoreAllCashiers,
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCashierListCache(ctx)
	logSuccess("Successfully restored all trashed cashiers", zap.Bool("success", success))
	return success, nil
}

func (s *cashierCommandService) DeleteAllCashierPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllCashierPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	success, err := s.cashierCommand.DeleteAllCashierPermanent(ctx)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](
			s.logger,
			cashier_errors.ErrFailedDeleteAllCashierPermanent,
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCashierListCache(ctx)
	logSuccess("Successfully deleted all trashed cashiers", zap.Bool("success", success))
	return success, nil
}
