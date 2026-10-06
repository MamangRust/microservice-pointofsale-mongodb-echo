package service

import (
	"context"

	mencache "github.com/MamangRust/microservice-point-of-sale-cashier/cache"
	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/outbox"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
)

type Service struct {
	CashierQuery   CashierQueryService
	CashierCommand CashierCommandService
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
		CashierQuery: NewCashierQueryService(&cashierQueryDeps{
			Cache:         deps.Mencache,
			CashierQuery:  deps.Repositories.CashierQuery,
			Logger:        deps.Logger,
			Observability: deps.Observability,
		}),
		CashierCommand: NewCashierCommandService(&cashierCommandDeps{
			Cache:          deps.Mencache,
			MerchantQuery:  deps.Repositories.MerchantQuery,
			UserQuery:      deps.Repositories.UserQuery,
			CashierCommand: deps.Repositories.CashierCommand,
			Outbox:         deps.Outbox,
			Logger:         deps.Logger,
			Observability:  deps.Observability,
		}),
	}
}
