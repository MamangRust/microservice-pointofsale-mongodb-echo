// Package role adapts the Role service gRPC API into the shared domain model.
package role

import (
	"context"

	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/role_errors"
)

// QueryRepository is the contract consumers depend on for role reads.
type QueryRepository interface {
	FindById(ctx context.Context, roleID int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
}

// Repository implements QueryRepository on top of the generated role service
// client. Role assignment lives in the user_role service, so this adapter only
// exposes the read path.
type Repository struct {
	client pbrole.RoleServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a role adapter. Passing zero options leaves the guard nil, which
// makes DependencyGuard.Call a plain passthrough.
func New(client pbrole.RoleServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *Repository) FindById(ctx context.Context, roleID int) (*models.Role, error) {
	var resp *pbrole.ApiResponseRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindByIdRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: int32(roleID)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, role_errors.ErrRoleNotFound.WithInternal(err)
	}
	return roleToModel(resp.Data), nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var resp *pbrole.ApiResponseRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindByNameRole(ctx, &pbrole.FindByNameRoleRequest{Name: name})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, role_errors.ErrRoleNotFound.WithInternal(err)
	}
	return roleToModel(resp.Data), nil
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
