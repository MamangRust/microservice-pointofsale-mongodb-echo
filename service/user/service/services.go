package service

import (
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	mencache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	"github.com/MamangRust/microservice-point-of-sale-user/repository"
)

type Service struct {
	UserQuery   UserQueryService
	UserCommand UserCommandService
}

type Deps struct {
	Mencache      mencache.Mencache
	Repositories  *repository.Repositories
	Hash          hash.HashPassword
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		UserQuery: NewUserQueryService(&userQueryDeps{
			Cache:         deps.Mencache,
			UserQuery:     deps.Repositories.UserQuery,
			Logger:        deps.Logger,
			Observability: deps.Observability,
		}),
		UserCommand: NewUserCommandService(&userCommandDeps{
			Cache:              deps.Mencache,
			UserQuery:          deps.Repositories.UserQuery,
			UserCommand:        deps.Repositories.UserCommand,
			RoleQuery:          deps.Repositories.Role,
			UserRoleRepository: deps.Repositories.UserRole,
			Logger:             deps.Logger,
			Hashing:            deps.Hash,
			Observability:      deps.Observability,
		}),
	}
}
