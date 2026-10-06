package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type RoleResult struct {
	RoleID     int32
	RoleName   string
	CreatedAt  *string
	UpdatedAt  *string
	DeletedAt  *string
	TotalCount int64
}

type RoleQueryRepository interface {
	FindAllRoles(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error)
	FindById(ctx context.Context, roleID int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindByUserId(ctx context.Context, userID int) ([]*models.Role, error)
	FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error)
	FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error)
}

type RoleCommandRepository interface {
	CreateRole(ctx context.Context, request *requests.CreateRoleRequest) (*models.Role, error)
	UpdateRole(ctx context.Context, request *requests.UpdateRoleRequest) (*models.Role, error)
	TrashedRole(ctx context.Context, roleID int) (*models.Role, error)
	RestoreRole(ctx context.Context, roleID int) (*models.Role, error)
	DeleteRolePermanent(ctx context.Context, roleID int) (bool, error)
	RestoreAllRole(ctx context.Context) (bool, error)
	DeleteAllRolePermanent(ctx context.Context) (bool, error)
}

type UserRoleCommandRepository interface {
	AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error
}
