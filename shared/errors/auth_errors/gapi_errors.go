package auth_errors

import (
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"

	"google.golang.org/grpc/codes"
)

var ErrGrpcLogin = response.NewGrpcError(
	"error",
	"login failed: invalid argument provided",
	int(codes.InvalidArgument),
)

var ErrGrpcGetMe = response.NewGrpcError(
	"error",
	"get user info failed: unauthenticated",
	int(codes.Unauthenticated),
)

var ErrGrpcRegisterToken = response.NewGrpcError(
	"error",
	"register failed: invalid argument",
	int(codes.InvalidArgument),
)
