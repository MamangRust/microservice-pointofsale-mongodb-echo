package service

import (
	"context"

	mencache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/outbox"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
)

type Service struct {
	OrderQuery   OrderQueryService
	OrderCommand OrderCommandService
}

type Deps struct {
	Ctx           context.Context
	Mencache      mencache.Mencache
	Repositories  *repository.Repositories
	Outbox        *outbox.OutboxService
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		OrderQuery: NewOrderQueryService(&orderQueryDeps{
			Cache:                deps.Mencache,
			OrderQueryRepository: deps.Repositories.OrderQuery,
			Logger:               deps.Logger,
			Observability:        deps.Observability,
		}),
		OrderCommand: NewOrderCommandService(&orderCommandDeps{
			Cache:                      deps.Mencache,
			CashierQueryRepository:     deps.Repositories.CashierQuery,
			OrderQueryRepository:       deps.Repositories.OrderQuery,
			OrderCommandRepository:     deps.Repositories.OrderCommand,
			OrderItemQueryRepository:   deps.Repositories.OrderItemQuery,
			OrderItemCommandRepository: deps.Repositories.OrderItemCommand,
			MerchantQueryRepository:    deps.Repositories.MerchantQuery,
			ProductQueryRepository:     deps.Repositories.ProductQuery,
			ProductCommandRepository:   deps.Repositories.ProductCommand,
			Outbox:                     deps.Outbox,
			Logger:                     deps.Logger,
			Observability:              deps.Observability,
		}),
	}
}
