package protomapper

import (
	authpb "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	userpb "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type authProtoMapper struct {
}

func NewAuthProtoMapper() *authProtoMapper {
	return &authProtoMapper{}
}

func (s *authProtoMapper) ToProtoResponseVerifyCode(status string, message string) *authpb.ApiResponseVerifyCode {
	return &authpb.ApiResponseVerifyCode{
		Status:  status,
		Message: message,
	}
}

func (s *authProtoMapper) ToProtoResponseForgotPassword(status string, message string) *authpb.ApiResponseForgotPassword {
	return &authpb.ApiResponseForgotPassword{
		Status:  status,
		Message: message,
	}
}

func (s *authProtoMapper) ToProtoResponseResetPassword(status string, message string) *authpb.ApiResponseResetPassword {
	return &authpb.ApiResponseResetPassword{
		Status:  status,
		Message: message,
	}
}

func (s *authProtoMapper) ToProtoResponseLogin(status string, message string, response *response.TokenResponse) *authpb.ApiResponseLogin {
	return &authpb.ApiResponseLogin{
		Status:  status,
		Message: message,
		Data: &authpb.TokenResponse{
			AccessToken:  response.AccessToken,
			RefreshToken: response.RefreshToken,
		},
	}
}

func (s *authProtoMapper) ToProtoResponseRegister(status string, message string, response *response.UserResponse) *authpb.ApiResponseRegister {
	return &authpb.ApiResponseRegister{
		Status:  status,
		Message: message,
		Data: &userpb.UserResponse{
			Id:        int32(response.ID),
			Firstname: response.FirstName,
			Lastname:  response.LastName,
			Email:     response.Email,
			CreatedAt: response.CreatedAt,
			UpdatedAt: response.UpdatedAt,
		},
	}
}

func (s *authProtoMapper) ToProtoResponseRefreshToken(status string, message string, response *response.TokenResponse) *authpb.ApiResponseRefreshToken {
	return &authpb.ApiResponseRefreshToken{
		Status:  status,
		Message: message,
		Data: &authpb.TokenResponse{
			AccessToken:  response.AccessToken,
			RefreshToken: response.RefreshToken,
		},
	}
}

func (s *authProtoMapper) ToProtoResponseGetMe(status string, message string, response *response.UserResponse) *authpb.ApiResponseGetMe {
	return &authpb.ApiResponseGetMe{
		Status:  status,
		Message: message,
		Data: &userpb.UserResponse{
			Id:        int32(response.ID),
			Firstname: response.FirstName,
			Lastname:  response.LastName,
			Email:     response.Email,
			CreatedAt: response.CreatedAt,
			UpdatedAt: response.UpdatedAt,
		},
	}
}
