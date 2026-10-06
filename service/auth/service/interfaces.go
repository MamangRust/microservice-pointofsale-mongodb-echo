package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type RegistrationService interface {
	Register(ctx context.Context, request *requests.RegisterRequest) (*models.User, error)
}

type LoginService interface {
	Login(ctx context.Context, request *requests.AuthRequest) (*response.TokenResponse, error)
}

type PasswordResetService interface {
	ForgotPassword(ctx context.Context, email string) (bool, error)

	ResetPassword(ctx context.Context, request *requests.CreateResetPasswordRequest) (bool, error)

	VerifyCode(ctx context.Context, code string) (bool, error)
}

type IdentifyService interface {
	RefreshToken(ctx context.Context, token string) (*response.TokenResponse, error)

	GetMe(ctx context.Context, token string) (*models.User, error)
}
