package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	AuthQuery   pbauth.AuthServiceServer
	AuthCommand pbauth.AuthServiceServer
}

func NewHandler(deps *Deps) *Handler {
	grpcHandler := NewAuthHandleGrpc(deps.Service, deps.Logger)
	return &Handler{
		AuthQuery:   grpcHandler,
		AuthCommand: grpcHandler,
	}
}
