package service

import (
	"context"

	mencache "github.com/MamangRust/microservice-point-of-sale-category/cache"
	"github.com/MamangRust/microservice-point-of-sale-category/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
)

type Service struct {
	CategoryQuery   CategoryQueryService
	CategoryCommand CategoryCommandService
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
		CategoryQuery: NewCategoryQueryService(&categoryQueryDeps{
			Cache:         deps.Mencache,
			CategoryQuery: deps.Repositories.CategoryQuery,
			Logger:        deps.Logger,
			Observability: deps.Observability,
		}),
		CategoryCommand: NewCategoryCommandService(&categoryCommandDeps{
			Cache:           deps.Mencache,
			CategoryQuery:   deps.Repositories.CategoryQuery,
			CategoryCommand: deps.Repositories.CategoryCommand,
			Logger:          deps.Logger,
			Observability:   deps.Observability,
		}),
	}
}
