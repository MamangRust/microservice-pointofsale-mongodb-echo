// @title Point of Sale REST API
// @version 1.0
// @description REST API Gateway (Echo) untuk monolith-pointofsale-grpc —
// @description REST → gRPC. Auth: login/register menghasilkan JWT Bearer token.
// @host localhost:5000
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/MamangRust/microservice-point-of-sale-apigateway/apps"
	_ "github.com/MamangRust/microservice-point-of-sale-apigateway/docs"
)

func main() {
	client, shutdown, err := apps.RunClient()
	if err != nil {
		log.Fatalf("failed to run client: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	client.Logger.Info("Gracefully shutting down...")
	shutdown()
}
