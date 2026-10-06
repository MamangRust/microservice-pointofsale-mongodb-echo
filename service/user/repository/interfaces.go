package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type UserResult struct {
	UserID     int32
	Firstname  string
	Lastname   string
	Email      string
	CreatedAt  *string
	UpdatedAt  *string
	DeletedAt  *string
	TotalCount int64
}

type UserQueryRepository interface {
	FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error)
	FindById(ctx context.Context, user_id int) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	// FindByEmailWithPassword returns the password hash so the auth service can
	// verify credentials. nil, nil when the user does not exist.
	FindByEmailWithPassword(ctx context.Context, email string) (*models.User, error)
	// FindByEmailAndVerify filters on is_verified = true; ErrUserNotFound when
	// no verified user matches.
	FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error)
	FindByVerificationCode(ctx context.Context, verification_code string) (*models.User, error)
	FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error)
}

type UserCommandRepository interface {
	CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*models.User, error)
	// CreateUserRecord inserts a user whose password arrives already hashed and
	// carries auth-managed fields (verification code, verified flag). Only the
	// auth service sends it.
	CreateUserRecord(ctx context.Context, request *requests.CreateUserRecordRequest) (*models.User, error)
	UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*models.User, error)
	UpdateUserIsVerified(ctx context.Context, user_id int, is_verified bool) (*models.User, error)
	UpdateUserPassword(ctx context.Context, user_id int, password string) (*models.User, error)
	TrashedUser(ctx context.Context, user_id int) (*models.User, error)
	RestoreUser(ctx context.Context, user_id int) (*models.User, error)
	DeleteUserPermanent(ctx context.Context, user_id int) (bool, error)
	RestoreAllUser(ctx context.Context) (bool, error)
	DeleteAllUserPermanent(ctx context.Context) (bool, error)
}

type RoleQueryRepository interface {
	FindByName(ctx context.Context, name string) (*models.Role, error)
}

// UserRoleRepository is the contract for role assignment, served by the role
// service through the pkg/adapter/role gRPC adapter.
type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error)

	RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error
}
