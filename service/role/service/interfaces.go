package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-role/repository"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type RoleQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllRoles) ([]*repository.RoleResult, *int, error)
	FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*repository.RoleResult, *int, error)
	FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*repository.RoleResult, *int, error)
	FindById(ctx context.Context, roleID int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindByUserId(ctx context.Context, id int) ([]*models.Role, error)
}

type RoleCommandService interface {
	CreateRole(ctx context.Context, request *requests.CreateRoleRequest) (*models.Role, error)
	UpdateRole(ctx context.Context, request *requests.UpdateRoleRequest) (*models.Role, error)
	TrashedRole(ctx context.Context, roleID int) (*models.Role, error)
	RestoreRole(ctx context.Context, roleID int) (*models.Role, error)
	DeleteRolePermanent(ctx context.Context, roleID int) (bool, error)
	RestoreAllRole(ctx context.Context) (bool, error)
	DeleteAllRolePermanent(ctx context.Context) (bool, error)
	AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error
}
