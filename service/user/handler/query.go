package handler

import (
	"context"
	"math"

	"github.com/MamangRust/microservice-point-of-sale-shared/convert"

	pbutils "github.com/MamangRust/microservice-point-of-sale-pb/common"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	user_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	"github.com/MamangRust/microservice-point-of-sale-user/repository"
	"github.com/MamangRust/microservice-point-of-sale-user/service"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type userQueryHandleGrpc struct {
	pb.UnimplementedUserServiceServer

	userQuery service.UserQueryService

	userCommandService service.UserCommandService
}

func NewUserHandleGrpc(query service.UserQueryService, command service.UserCommandService) *userQueryHandleGrpc {
	return &userQueryHandleGrpc{
		userQuery:          query,
		userCommandService: command,
	}
}

func (s *userQueryHandleGrpc) FindAll(ctx context.Context, request *pb.FindAllUserRequest) (*pb.ApiResponsePaginationUser, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	userResponses := make([]*pb.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.StrVal(user.CreatedAt),
			UpdatedAt: convert.StrVal(user.UpdatedAt),
		}
	}

	return &pb.ApiResponsePaginationUser{
		Status:     "success",
		Message:    "Successfully fetched users",
		Data:       userResponses,
		Pagination: paginationMeta,
	}, nil
}

func (s *userQueryHandleGrpc) FindById(ctx context.Context, request *pb.FindByIdUserRequest) (*pb.ApiResponseUser, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, user_errors.ErrUserNotFound
	}

	user, err := s.userQuery.FindByID(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully fetched user",
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

func (s *userQueryHandleGrpc) FindByActive(ctx context.Context, request *pb.FindAllUserRequest) (*pb.ApiResponsePaginationUserDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQuery.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	userResponses := make([]*pb.UserResponseDeleteAt, len(users))
	for i, user := range users {
		userResponses[i] = &pb.UserResponseDeleteAt{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.StrVal(user.CreatedAt),
			UpdatedAt: convert.StrVal(user.UpdatedAt),
			DeletedAt: convert.StrValToWrappers(user.DeletedAt),
		}
	}

	return &pb.ApiResponsePaginationUserDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active users",
		Data:       userResponses,
		Pagination: paginationMeta,
	}, nil
}

func (s *userQueryHandleGrpc) FindByTrashed(ctx context.Context, request *pb.FindAllUserRequest) (*pb.ApiResponsePaginationUserDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	userResponses := make([]*pb.UserResponseDeleteAt, len(users))
	for i, user := range users {
		userResponses[i] = &pb.UserResponseDeleteAt{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.StrVal(user.CreatedAt),
			UpdatedAt: convert.StrVal(user.UpdatedAt),
			DeletedAt: convert.StrValToWrappers(user.DeletedAt),
		}
	}

	return &pb.ApiResponsePaginationUserDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed users",
		Data:       userResponses,
		Pagination: paginationMeta,
	}, nil
}

func (s *userQueryHandleGrpc) FindByEmail(ctx context.Context, request *pb.FindByEmailUserRequest) (*pb.ApiResponseUserWithPassword, error) {
	email := request.GetEmail()
	if email == "" {
		return nil, user_errors.ErrUserNotFound
	}

	user, err := s.userQuery.FindByEmailWithPassword(ctx, email)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	if user == nil {
		return nil, user_errors.ErrUserNotFound
	}

	return &pb.ApiResponseUserWithPassword{
		Status:  "success",
		Message: "Successfully fetched user",
		Data:    mapUserWithPasswordModel(user),
	}, nil
}

func (s *userQueryHandleGrpc) FindByEmailAndVerify(ctx context.Context, request *pb.FindByEmailUserRequest) (*pb.ApiResponseUserWithPassword, error) {
	email := request.GetEmail()
	if email == "" {
		return nil, user_errors.ErrUserNotFound
	}

	user, err := s.userQuery.FindByEmailAndVerify(ctx, email)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUserWithPassword{
		Status:  "success",
		Message: "Successfully fetched verified user",
		Data:    mapUserWithPasswordModel(user),
	}, nil
}

func (s *userQueryHandleGrpc) FindByVerificationCode(ctx context.Context, request *pb.FindByVerificationCodeUserRequest) (*pb.ApiResponseUser, error) {
	code := request.GetVerificationCode()
	if code == "" {
		return nil, user_errors.ErrUserNotFound
	}

	user, err := s.userQuery.FindByVerificationCode(ctx, code)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully fetched user by verification code",
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

func mapUserWithPasswordModel(user *models.User) *pb.UserResponseWithPassword {
	if user == nil {
		return nil
	}
	return &pb.UserResponseWithPassword{
		Id:        user.UserID,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: convert.FormatTimePtr(user.CreatedAt),
		UpdatedAt: convert.FormatTimePtr(user.UpdatedAt),
	}
}

func mapUserResults(users []*repository.UserResult) []*pb.UserResponse {
	var res []*pb.UserResponse
	for _, u := range users {
		res = append(res, &pb.UserResponse{
			Id:        u.UserID,
			Firstname: u.Firstname,
			Lastname:  u.Lastname,
			Email:     u.Email,
			CreatedAt: convert.StrVal(u.CreatedAt),
			UpdatedAt: convert.StrVal(u.UpdatedAt),
		})
	}
	return res
}

func mapUserResultsDeleteAt(users []*repository.UserResult) []*pb.UserResponseDeleteAt {
	var res []*pb.UserResponseDeleteAt
	for _, u := range users {
		var deletedAt *wrapperspb.StringValue
		if u.DeletedAt != nil {
			deletedAt = wrapperspb.String(*u.DeletedAt)
		}
		res = append(res, &pb.UserResponseDeleteAt{
			Id:        u.UserID,
			Firstname: u.Firstname,
			Lastname:  u.Lastname,
			Email:     u.Email,
			CreatedAt: convert.StrVal(u.CreatedAt),
			UpdatedAt: convert.StrVal(u.UpdatedAt),
			DeletedAt: deletedAt,
		})
	}
	return res
}
