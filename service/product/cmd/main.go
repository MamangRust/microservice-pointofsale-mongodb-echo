package main

import (
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/server"
	"github.com/MamangRust/microservice-point-of-sale-product/apps"
	"github.com/spf13/viper"
)

func main() {
	port := viper.GetInt("GRPC_PRODUCT_ADDR")
	if port == 0 {
		port = 50059
	}

	srv, err := apps.NewServer(&server.Config{
		ServiceName:    "product-service",
		DBName:         database.CatalogDB,
		ServiceVersion: "1.0.0",
		Environment:    "production",
		OtelEndpoint:   "otel-collector:4317",
		Port:           port,
	})
	if err != nil {
		panic(err)
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
