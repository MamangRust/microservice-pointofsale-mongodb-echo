// Package user adapts the User service gRPC API into the shared domain model.
package user

import (
	"context"

	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
)

// QueryRepository is the contract consumers depend on for user reads.
type QueryRepository interface {
	FindById(ctx context.Context, id int) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error)
	FindByVerificationCode(ctx context.Context, code string) (*models.User, error)
}

// CommandRepository is the contract consumers depend on for user writes on
// behalf of the auth service.
type CommandRepository interface {
	CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error)
	UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error)
	UpdateUserPassword(ctx context.Context, userID int, password string) (*models.User, error)
	DeleteUserPermanent(ctx context.Context, userID int) error
}

// Repository implements QueryRepository and CommandRepository on top of the
// generated user service client.
type Repository struct {
	client pbuser.UserServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a user adapter. Passing zero options leaves the guard nil, which
// makes DependencyGuard.Call a plain passthrough.
func New(client pbuser.UserServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *Repository) FindById(ctx context.Context, id int) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindById(ctx, &pbuser.FindByIdUserRequest{Id: int32(id)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrInternal
	}

	u := resp.Data
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		CreatedAt: convert.TimePtr(u.CreatedAt),
		UpdatedAt: convert.TimePtr(u.UpdatedAt),
	}, nil
}

// FindByEmail returns the user with the password hash. A missing user comes
// back as ErrUserNotFound so consumers can distinguish it from transport
// failures via errors.Is.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var resp *pbuser.ApiResponseUserWithPassword
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindByEmail(ctx, &pbuser.FindByEmailUserRequest{Email: email})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, user_errors.ErrUserNotFound.WithInternal(err)
	}
	return userWithPasswordToModel(resp.Data), nil
}

// FindByEmailAndVerify mirrors the legacy auth repo behaviour: lookup by email
// restricted to verified users, used for credential verification on login.
func (r *Repository) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	var resp *pbuser.ApiResponseUserWithPassword
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindByEmailAndVerify(ctx, &pbuser.FindByEmailUserRequest{Email: email})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, user_errors.ErrUserNotFound.WithInternal(err)
	}
	return userWithPasswordToModel(resp.Data), nil
}

func (r *Repository) FindByVerificationCode(ctx context.Context, code string) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindByVerificationCode(ctx, &pbuser.FindByVerificationCodeUserRequest{VerificationCode: code})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, user_errors.ErrUserNotFound.WithInternal(err)
	}

	u := resp.Data
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		CreatedAt: convert.TimePtr(u.CreatedAt),
		UpdatedAt: convert.TimePtr(u.UpdatedAt),
	}, nil
}

// CreateUser registers a user through the User service. The password must
// already be hashed by the caller; the verification fields pass through as
// auth-managed values.
func (r *Repository) CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.CreateUserRecord(ctx, &pbuser.CreateUserRecordRequest{
			Firstname:        request.FirstName,
			Lastname:         request.LastName,
			Email:            request.Email,
			Password:         request.Password,
			VerificationCode: request.VerifiedCode,
			IsVerified:       request.IsVerified,
		})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, user_errors.ErrCreateUser.WithInternal(err)
	}
	return userToModel(resp.Data), nil
}

func (r *Repository) UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.UpdateUserIsVerified(ctx, &pbuser.UpdateUserIsVerifiedUserRequest{
			Id:         int32(userID),
			IsVerified: isVerified,
		})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, user_errors.ErrUpdateUserVerificationCode.WithInternal(err)
	}
	return userToModel(resp.Data), nil
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID int, password string) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.UpdateUserPassword(ctx, &pbuser.UpdateUserPasswordUserRequest{
			Id:       int32(userID),
			Password: password,
		})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, user_errors.ErrUpdateUserPassword.WithInternal(err)
	}
	return userToModel(resp.Data), nil
}

// DeleteUserPermanent removes a user for good. The User service only purges rows
// that are already soft-deleted, so this first trashes the user and then purges
// it. That invariant is hidden here because callers (auth's register
// compensation) just want the user gone, whatever its current state.
func (r *Repository) DeleteUserPermanent(ctx context.Context, userID int) error {
	// Best effort: a user that is already trashed returns an error here, which
	// must not stop the purge below.
	_ = r.guard.Call(ctx, func(ctx context.Context) error {
		_, callErr := r.client.TrashedUser(ctx, &pbuser.FindByIdUserRequest{Id: int32(userID)})
		return callErr
	})

	var resp *pbuser.ApiResponseUserDelete
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.DeleteUserPermanent(ctx, &pbuser.FindByIdUserRequest{Id: int32(userID)})
		return callErr
	})
	if err != nil || resp == nil {
		return user_errors.ErrDeleteUserPermanent.WithInternal(err)
	}
	return nil
}

func userToModel(u *pbuser.UserResponse) *models.User {
	if u == nil {
		return nil
	}
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		CreatedAt: convert.TimePtr(u.CreatedAt),
		UpdatedAt: convert.TimePtr(u.UpdatedAt),
	}
}

func userWithPasswordToModel(u *pbuser.UserResponseWithPassword) *models.User {
	if u == nil {
		return nil
	}
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: convert.TimePtr(u.CreatedAt),
		UpdatedAt: convert.TimePtr(u.UpdatedAt),
	}
}
