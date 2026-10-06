package handler

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	traceunic "github.com/MamangRust/microservice-point-of-sale-pkg/trace_unic"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"go.uber.org/zap"
)

type authHandleGrpc struct {
	pbauth.UnimplementedAuthServiceServer
	registerService      service.RegistrationService
	loginService         service.LoginService
	passwordResetService service.PasswordResetService
	identifyService      service.IdentifyService
	logger               logger.LoggerInterface
}

func NewAuthHandleGrpc(authService *service.Service, logger logger.LoggerInterface) *authHandleGrpc {
	return &authHandleGrpc{
		registerService:      authService.Register,
		loginService:         authService.Login,
		passwordResetService: authService.PasswordReset,
		identifyService:      authService.Identify,
		logger:               logger,
	}
}

func fmtAuthTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func mapAuthUserResponse(user *models.User) *pbuser.UserResponse {
	if user == nil {
		return nil
	}
	return &pbuser.UserResponse{
		Id:        user.UserID,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: fmtAuthTime(user.CreatedAt),
		UpdatedAt: fmtAuthTime(user.UpdatedAt),
	}
}

func (s *authHandleGrpc) VerifyCode(ctx context.Context, req *pbauth.VerifyCodeRequest) (*pbauth.ApiResponseVerifyCode, error) {
	s.logger.Info("VerifyCode called", zap.Bool("verification_code_present", req.Code != ""))

	_, err := s.passwordResetService.VerifyCode(ctx, req.Code)
	if err != nil {
		traceId := traceunic.GenerateTraceID("VERIFY_CODE_FAILED")
		s.logger.Error("VerifyCode failed", zap.String("traceId", traceId), zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("VerifyCode success", zap.Bool("verification_code_present", req.Code != ""))

	return &pbauth.ApiResponseVerifyCode{
		Status:  "success",
		Message: "Verification successfully",
	}, nil
}

func (s *authHandleGrpc) ForgotPassword(ctx context.Context, req *pbauth.ForgotPasswordRequest) (*pbauth.ApiResponseForgotPassword, error) {
	s.logger.Info("ForgotPassword called", zap.String("email", req.Email))

	_, err := s.passwordResetService.ForgotPassword(ctx, req.Email)
	if err != nil {
		traceId := traceunic.GenerateTraceID("FORGOT_PASSWORD_FAILED")
		s.logger.Error("ForgotPassword failed", zap.String("traceId", traceId), zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("ForgotPassword successful", zap.Bool("success", true))

	return &pbauth.ApiResponseForgotPassword{
		Status:  "success",
		Message: "ForgotPassword successful",
	}, nil
}

func (s *authHandleGrpc) ResetPassword(ctx context.Context, req *pbauth.ResetPasswordRequest) (*pbauth.ApiResponseResetPassword, error) {
	s.logger.Info("ResetPassword called", zap.Bool("reset_token_present", req.ResetToken != ""))

	_, err := s.passwordResetService.ResetPassword(ctx, &requests.CreateResetPasswordRequest{
		ResetToken:      req.ResetToken,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	})
	if err != nil {
		traceId := traceunic.GenerateTraceID("RESET_PASSWORD_FAILED")
		s.logger.Error("ResetPassword failed", zap.String("traceId", traceId), zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("ResetPassword successful", zap.Bool("success", true))

	return &pbauth.ApiResponseResetPassword{
		Status:  "success",
		Message: "Reset password successful",
	}, nil
}

func (s *authHandleGrpc) LoginUser(ctx context.Context, req *pbauth.LoginRequest) (*pbauth.ApiResponseLogin, error) {
	s.logger.Info("LoginUser called", zap.String("email", req.Email))

	request := &requests.AuthRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	res, err := s.loginService.Login(ctx, request)
	if err != nil {
		traceId := traceunic.GenerateTraceID("LOGIN_FAILED")
		s.logger.Error("LoginUser failed", zap.String("traceId", traceId), zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("LoginUser successful", zap.Bool("success", true))

	return &pbauth.ApiResponseLogin{
		Status:  "success",
		Message: "LoginUser successfull",
		Data: &pbauth.TokenResponse{
			AccessToken:  res.AccessToken,
			RefreshToken: res.RefreshToken,
		},
	}, nil
}

func (s *authHandleGrpc) RefreshToken(ctx context.Context, req *pbauth.RefreshTokenRequest) (*pbauth.ApiResponseRefreshToken, error) {
	s.logger.Info("RefreshToken called")

	res, err := s.identifyService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		traceId := traceunic.GenerateTraceID("REFRESH_TOKEN_FAILED")
		s.logger.Error("RefreshToken failed", zap.String("traceId", traceId), zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RefreshToken successful", zap.Bool("success", true))

	return &pbauth.ApiResponseRefreshToken{
		Status:  "success",
		Message: "Refresh token successful",
		Data: &pbauth.TokenResponse{
			AccessToken:  res.AccessToken,
			RefreshToken: req.RefreshToken,
		},
	}, nil
}

func (s *authHandleGrpc) GetMe(ctx context.Context, req *pbauth.GetMeRequest) (*pbauth.ApiResponseGetMe, error) {
	s.logger.Info("GetMe called")

	res, err := s.identifyService.GetMe(ctx, req.AccessToken)
	if err != nil {
		traceId := traceunic.GenerateTraceID("GET_ME_FAILED")
		s.logger.Error("GetMe failed", zap.String("traceId", traceId), zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("GetMe successful", zap.Bool("success", true))

	return &pbauth.ApiResponseGetMe{
		Status:  "success",
		Message: "Get me successfully",
		Data:    mapAuthUserResponse(res),
	}, nil
}

func (s *authHandleGrpc) RegisterUser(ctx context.Context, req *pbauth.RegisterRequest) (*pbauth.ApiResponseRegister, error) {
	s.logger.Info("RegisterUser called", zap.String("email", req.Email))

	request := &requests.RegisterRequest{
		FirstName:       req.Firstname,
		LastName:        req.Lastname,
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}

	res, err := s.registerService.Register(ctx, request)
	if err != nil {
		traceId := traceunic.GenerateTraceID("REGISTER_FAILED")
		s.logger.Error("RegisterUser failed", zap.String("traceId", traceId), zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RegisterUser successful", zap.Bool("success", true))

	return &pbauth.ApiResponseRegister{
		Status:  "success",
		Message: "RegisterUser successful",
		Data:    mapAuthUserResponse(res),
	}, nil
}
