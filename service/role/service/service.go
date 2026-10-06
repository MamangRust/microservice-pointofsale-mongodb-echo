package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	mencache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	"github.com/MamangRust/microservice-point-of-sale-role/repository"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
)

type Service struct {
	RoleQuery   RoleQueryService
	RoleCommand RoleCommandService
}

type Deps struct {
	Ctx           context.Context
	Mencache      mencache.Mencache
	Repositories  *repository.Repositories
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		RoleQuery:   NewRoleQueryService(deps.Mencache, deps.Repositories.RoleQuery, deps.Logger, deps.Observability),
		RoleCommand: NewRoleCommandService(deps.Mencache, deps.Repositories.RoleCommand, deps.Repositories.UserRoleCommand, deps.Logger, deps.Observability),
	}
}
