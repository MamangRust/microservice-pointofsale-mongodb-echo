package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-category/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CategoryQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllCategory) ([]*repository.CategoryResult, *int, error)
	FindById(ctx context.Context, category_id int) (*models.Category, error)
	FindByIds(ctx context.Context, category_ids []int) ([]*models.Category, error)
	FindByName(ctx context.Context, name string) (*models.Category, error)
	FindByActive(ctx context.Context, req *requests.FindAllCategory) ([]*repository.CategoryResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*repository.CategoryResultDeleteAt, *int, error)
}

type CategoryCommandService interface {
	CreateCategory(ctx context.Context, req *requests.CreateCategoryRequest) (*models.Category, error)
	UpdateCategory(ctx context.Context, req *requests.UpdateCategoryRequest) (*models.Category, error)
	TrashedCategory(ctx context.Context, category_id int) (*models.Category, error)
	RestoreCategory(ctx context.Context, categoryID int) (*models.Category, error)
	DeleteCategoryPermanent(ctx context.Context, categoryID int) (bool, error)
	RestoreAllCategories(ctx context.Context) (bool, error)
	DeleteAllCategoriesPermanent(ctx context.Context) (bool, error)
}
