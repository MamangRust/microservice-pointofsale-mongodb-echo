// Package user_role adapts the UserRole service gRPC API into the shared domain
// model. It owns the only place that talks to pb/user_role's UserRoleService.
package user_role

import (
	"context"

	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	userrole_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/user_role_errors"
)

// QueryRepository is the read path consumers use to resolve a user's roles.
type QueryRepository interface {
	FindByUserId(ctx context.Context, userID int) ([]*models.Role, error)
}

// CommandRepository is the write path consumers use to (un)assign roles.
type CommandRepository interface {
	AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error
}

// Repository implements QueryRepository and CommandRepository on top of the
// generated user-role service client.
type Repository struct {
	client pbuserrole.UserRoleServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a user-role adapter. Passing zero options leaves the guard nil,
// which makes DependencyGuard.Call a plain passthrough.
func New(client pbuserrole.UserRoleServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

// FindByUserId implements QueryRepository. The user-role service returns the
// resolved roles, so we map the role responses straight into the domain model.
func (r *Repository) FindByUserId(ctx context.Context, userID int) ([]*models.Role, error) {
	var resp *pbrole.ApiResponsesRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindByUserId(ctx, &pbuserrole.FindByIdUserRoleRequest{UserId: int32(userID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	roles := make([]*models.Role, 0, len(resp.Data))
	for _, item := range resp.Data {
		if item == nil {
			continue
		}
		roles = append(roles, roleToModel(item))
	}
	return roles, nil
}

// AssignRoleToUser implements CommandRepository.
func (r *Repository) AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	var resp *pbuserrole.ApiResponseUserRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.AssignRoleToUser(ctx, &pbuserrole.AssignRoleToUserRequest{
			UserId: int32(req.UserId),
			RoleId: int32(req.RoleId),
		})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, userrole_errors.ErrAssignRoleToUser.WithInternal(err)
	}
	return userRoleToModel(resp.Data), nil
}

// RemoveRoleFromUser implements CommandRepository.
func (r *Repository) RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error {
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		_, callErr := r.client.RemoveRoleFromUser(ctx, &pbuserrole.RemoveRoleFromUserRequest{
			UserId: int32(req.UserId),
			RoleId: int32(req.RoleId),
		})
		return callErr
	})
	if err != nil {
		return userrole_errors.ErrRemoveRole.WithInternal(err)
	}
	return nil
}

func roleToModel(role *pbrole.RoleResponse) *models.Role {
	if role == nil {
		return nil
	}
	return &models.Role{
		RoleID:    role.Id,
		RoleName:  role.Name,
		CreatedAt: convert.TimePtr(role.CreatedAt),
		UpdatedAt: convert.TimePtr(role.UpdatedAt),
	}
}

func userRoleToModel(ur *pbuserrole.UserRoleResponse) *models.UserRole {
	if ur == nil {
		return nil
	}
	return &models.UserRole{
		UserRoleID: ur.UserRoleId,
		UserID:     ur.UserId,
		RoleID:     ur.RoleId,
	}
}
