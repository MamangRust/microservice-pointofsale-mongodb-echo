// Package category adapts the Category service gRPC API into the shared domain
// model.
package category

import (
	"context"

	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
)

// Repository reads categories from the Category service over gRPC.
type Repository interface {
	FindById(ctx context.Context, id int) (*models.Category, error)
	FindByIds(ctx context.Context, ids []int) ([]*models.Category, error)
	FindByName(ctx context.Context, name string) (*models.Category, error)
}

type repository struct {
	client pbcategory.CategoryServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a category adapter. Passing zero options leaves the guard nil,
// which makes DependencyGuard.Call a plain passthrough.
func New(client pbcategory.CategoryServiceClient, opts ...adapter.GuardOption) Repository {
	r := &repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *repository) FindById(ctx context.Context, id int) (*models.Category, error) {
	var resp *pbcategory.ApiResponseCategory
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindById(ctx, &pbcategory.FindByIdCategoryRequest{Id: int32(id)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, category_errors.ErrFindById
	}

	return mapCategory(resp.Data), nil
}

// FindByIds resolves many categories in a single call. Missing or soft-deleted
// ids are simply absent from the result; callers must tolerate that.
func (r *repository) FindByIds(ctx context.Context, ids []int) ([]*models.Category, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	req := &pbcategory.FindByIdsCategoryRequest{Ids: make([]int32, 0, len(ids))}
	for _, id := range ids {
		req.Ids = append(req.Ids, int32(id))
	}

	var resp *pbcategory.ApiResponsesCategory
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindByIds(ctx, req)
		return callErr
	})
	if err != nil || resp == nil {
		return nil, category_errors.ErrFindByIds
	}

	categories := make([]*models.Category, 0, len(resp.Data))
	for _, c := range resp.Data {
		if c == nil {
			continue
		}
		categories = append(categories, mapCategory(c))
	}
	return categories, nil
}

func (r *repository) FindByName(ctx context.Context, name string) (*models.Category, error) {
	var resp *pbcategory.ApiResponseCategory
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindByName(ctx, &pbcategory.FindByNameCategoryRequest{Name: name})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, category_errors.ErrFindByName
	}

	return mapCategory(resp.Data), nil
}

func mapCategory(c *pbcategory.CategoryResponse) *models.Category {
	if c == nil {
		return nil
	}
	return &models.Category{
		CategoryID:   c.Id,
		Name:         c.Name,
		Description:  convert.NullableString(c.Description),
		SlugCategory: convert.NullableString(c.SlugCategory),
		CreatedAt:    convert.TimePtr(c.CreatedAt),
		UpdatedAt:    convert.TimePtr(c.UpdatedAt),
	}
}
