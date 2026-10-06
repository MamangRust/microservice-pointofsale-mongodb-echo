package handler

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-shared/convert"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	user_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *userQueryHandleGrpc) Create(ctx context.Context, request *pb.CreateUserRequest) (*pb.ApiResponseUser, error) {
	req := &requests.CreateUserRequest{
		FirstName:       request.GetFirstname(),
		LastName:        request.GetLastname(),
		Email:           request.GetEmail(),
		Password:        request.GetPassword(),
		ConfirmPassword: request.GetConfirmPassword(),
	}

	user, err := s.userCommandService.CreateUser(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully created user",
		Data: &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimePtr(user.CreatedAt),
			UpdatedAt: convert.FormatTimePtr(user.UpdatedAt),
		},
	}, nil
}

func (s *userQueryHandleGrpc) Update(ctx context.Context, request *pb.UpdateUserRequest) (*pb.ApiResponseUser, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, user_errors.ErrUserNotFound
	}

	req := &requests.UpdateUserRequest{
		UserID:          &id,
		FirstName:       request.GetFirstname(),
		LastName:        request.GetLastname(),
		Email:           request.GetEmail(),
		Password:        request.GetPassword(),
		ConfirmPassword: request.GetConfirmPassword(),
	}

	user, err := s.userCommandService.UpdateUser(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully updated user",
		Data: &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimePtr(user.CreatedAt),
			UpdatedAt: convert.FormatTimePtr(user.UpdatedAt),
		},
	}, nil
}

func (s *userQueryHandleGrpc) CreateUserRecord(ctx context.Context, request *pb.CreateUserRecordRequest) (*pb.ApiResponseUser, error) {
	req := &requests.CreateUserRecordRequest{
		FirstName:        request.GetFirstname(),
		LastName:         request.GetLastname(),
		Email:            request.GetEmail(),
		Password:         request.GetPassword(),
		VerificationCode: request.GetVerificationCode(),
		IsVerified:       request.GetIsVerified(),
	}

	user, err := s.userCommandService.CreateUserRecord(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully created user record",
		Data: &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimePtr(user.CreatedAt),
			UpdatedAt: convert.FormatTimePtr(user.UpdatedAt),
		},
	}, nil
}

func (s *userQueryHandleGrpc) UpdateUserIsVerified(ctx context.Context, request *pb.UpdateUserIsVerifiedUserRequest) (*pb.ApiResponseUser, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, user_errors.ErrUserNotFound
	}

	user, err := s.userCommandService.UpdateUserIsVerified(ctx, id, request.GetIsVerified())
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully updated user verification",
		Data: &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimePtr(user.CreatedAt),
			UpdatedAt: convert.FormatTimePtr(user.UpdatedAt),
		},
	}, nil
}

func (s *userQueryHandleGrpc) UpdateUserPassword(ctx context.Context, request *pb.UpdateUserPasswordUserRequest) (*pb.ApiResponseUser, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, user_errors.ErrUserNotFound
	}

	user, err := s.userCommandService.UpdateUserPassword(ctx, id, request.GetPassword())
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully updated user password",
		Data: &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimePtr(user.CreatedAt),
			UpdatedAt: convert.FormatTimePtr(user.UpdatedAt),
		},
	}, nil
}

func (s *userQueryHandleGrpc) TrashedUser(ctx context.Context, request *pb.FindByIdUserRequest) (*pb.ApiResponseUserDeleteAt, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, user_errors.ErrUserNotFound
	}

	user, err := s.userCommandService.TrashedUser(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var deletedAt *wrapperspb.StringValue
	if user.DeletedAt != nil {
		deletedAt = &wrapperspb.StringValue{Value: user.DeletedAt.Format(time.RFC3339)}
	}

	return &pb.ApiResponseUserDeleteAt{
		Status:  "success",
		Message: "Successfully trashed user",
		Data: &pb.UserResponseDeleteAt{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimePtr(user.CreatedAt),
			UpdatedAt: convert.FormatTimePtr(user.UpdatedAt),
			DeletedAt: deletedAt,
		},
	}, nil
}

func (s *userQueryHandleGrpc) RestoreUser(ctx context.Context, request *pb.FindByIdUserRequest) (*pb.ApiResponseUserDeleteAt, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, user_errors.ErrUserNotFound
	}

	user, err := s.userCommandService.RestoreUser(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUserDeleteAt{
		Status:  "success",
		Message: "Successfully restored user",
		Data: &pb.UserResponseDeleteAt{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimePtr(user.CreatedAt),
			UpdatedAt: convert.FormatTimePtr(user.UpdatedAt),
			DeletedAt: nil,
		},
	}, nil
}

func (s *userQueryHandleGrpc) DeleteUserPermanent(ctx context.Context, request *pb.FindByIdUserRequest) (*pb.ApiResponseUserDelete, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, user_errors.ErrUserNotFound
	}

	_, err := s.userCommandService.DeleteUserPermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUserDelete{
		Status:  "success",
		Message: "Successfully deleted user permanently",
	}, nil
}

func (s *userQueryHandleGrpc) RestoreAllUser(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseUserAll, error) {
	_, err := s.userCommandService.RestoreAllUser(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUserAll{
		Status:  "success",
		Message: "Successfully restored all users",
	}, nil
}

func (s *userQueryHandleGrpc) DeleteAllUserPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseUserAll, error) {
	_, err := s.userCommandService.DeleteAllUserPermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUserAll{
		Status:  "success",
		Message: "Successfully deleted all users permanently",
	}, nil
}
