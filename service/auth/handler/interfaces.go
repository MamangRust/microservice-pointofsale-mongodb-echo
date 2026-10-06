package handler

import (
	"context"

	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
)

type AuthHandleGrpc interface {
	pbauth.AuthServiceServer
	VerifyCode(ctx context.Context, req *pbauth.VerifyCodeRequest) (*pbauth.ApiResponseVerifyCode, error)
	ForgotPassword(ctx context.Context, req *pbauth.ForgotPasswordRequest) (*pbauth.ApiResponseForgotPassword, error)
	ResetPassword(ctx context.Context, req *pbauth.ResetPasswordRequest) (*pbauth.ApiResponseResetPassword, error)
	LoginUser(ctx context.Context, req *pbauth.LoginRequest) (*pbauth.ApiResponseLogin, error)
	RefreshToken(ctx context.Context, req *pbauth.RefreshTokenRequest) (*pbauth.ApiResponseRefreshToken, error)
	GetMe(ctx context.Context, req *pbauth.GetMeRequest) (*pbauth.ApiResponseGetMe, error)
	RegisterUser(ctx context.Context, req *pbauth.RegisterRequest) (*pbauth.ApiResponseRegister, error)
}
