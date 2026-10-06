// Package merchant adapts the Merchant service gRPC API into the shared
// domain model used by the consumers of merchants (cashier, order,
// transaction, product).
package merchant

import (
	"context"

	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_errors"
)

// Repository reads merchants from the Merchant service over gRPC.
type Repository interface {
	FindById(ctx context.Context, id int) (*models.Merchant, error)
}

type repository struct {
	client pbmerchant.MerchantServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a merchant adapter. Passing zero options leaves the guard nil,
// which makes DependencyGuard.Call a plain passthrough.
func New(client pbmerchant.MerchantServiceClient, opts ...adapter.GuardOption) Repository {
	r := &repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *repository) FindById(ctx context.Context, id int) (*models.Merchant, error) {
	var resp *pbmerchant.ApiResponseMerchant
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindById(ctx, &pbmerchant.FindByIdMerchantRequest{Id: int32(id)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, merchant_errors.ErrFindById
	}

	m := resp.Data
	return &models.Merchant{
		MerchantID:   m.Id,
		UserID:       m.UserId,
		Name:         m.Name,
		Description:  convert.NullableString(m.Description),
		Address:      convert.NullableString(m.Address),
		ContactEmail: convert.NullableString(m.ContactEmail),
		ContactPhone: convert.NullableString(m.ContactPhone),
		Status:       m.Status,
		CreatedAt:    convert.TimePtr(m.CreatedAt),
		UpdatedAt:    convert.TimePtr(m.UpdatedAt),
	}, nil
}
