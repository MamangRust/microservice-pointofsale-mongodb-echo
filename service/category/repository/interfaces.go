package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// Local result types for stats queries

type CategoryResult struct {
	CategoryID   int32
	Name         string
	Description  *string
	SlugCategory *string
	CreatedAt    *string
	UpdatedAt    *string
	TotalCount   int64
}

type CategoryResultDeleteAt struct {
	CategoryID   int32
	Name         string
	Description  *string
	SlugCategory *string
	CreatedAt    *string
	UpdatedAt    *string
	DeletedAt    *string
	TotalCount   int64
}

type CategoryQueryRepository interface {
	FindAllCategory(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResult, *int, error)
	FindById(ctx context.Context, category_id int) (*models.Category, error)
	FindByIds(ctx context.Context, category_ids []int) ([]*models.Category, error)
	FindByNameAndId(ctx context.Context, req *requests.CategoryNameAndId) (*models.Category, error)
	FindByName(ctx context.Context, name string) (*models.Category, error)
	FindByIdTrashed(ctx context.Context, category_id int) (*models.Category, error)
	FindByActive(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResultDeleteAt, *int, error)
}

type CategoryCommandRepository interface {
	CreateCategory(ctx context.Context, request *requests.CreateCategoryRequest) (*models.Category, error)
	UpdateCategory(ctx context.Context, request *requests.UpdateCategoryRequest) (*models.Category, error)
	TrashedCategory(ctx context.Context, category_id int) (*models.Category, error)
	RestoreCategory(ctx context.Context, category_id int) (*models.Category, error)
	DeleteCategoryPermanently(ctx context.Context, category_id int) (bool, error)
	RestoreAllCategories(ctx context.Context) (bool, error)
	DeleteAllPermanentCategories(ctx context.Context) (bool, error)
}
