package protomapper

import (
	commonpb "github.com/MamangRust/microservice-point-of-sale-pb/common"
	userpb "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type userProtoMapper struct {
}

func NewUserProtoMapper() *userProtoMapper {
	return &userProtoMapper{}
}

func (u *userProtoMapper) ToProtoResponseUser(status string, message string, pbResponse *response.UserResponse) *userpb.ApiResponseUser {
	return &userpb.ApiResponseUser{
		Status:  status,
		Message: message,
		Data:    u.mapResponseUser(pbResponse),
	}
}

func (u *userProtoMapper) ToProtoResponseUserDeleteAt(status string, message string, pbResponse *response.UserResponseDeleteAt) *userpb.ApiResponseUserDeleteAt {
	return &userpb.ApiResponseUserDeleteAt{
		Status:  status,
		Message: message,
		Data:    u.mapResponseUserDeleteAt(pbResponse),
	}
}

func (u *userProtoMapper) ToProtoResponsesUser(status string, message string, pbResponse []*response.UserResponse) *userpb.ApiResponsesUser {
	return &userpb.ApiResponsesUser{
		Status:  status,
		Message: message,
		Data:    u.mapResponsesUser(pbResponse),
	}
}

func (u *userProtoMapper) ToProtoResponseUserDelete(status string, message string) *userpb.ApiResponseUserDelete {
	return &userpb.ApiResponseUserDelete{
		Status:  status,
		Message: message,
	}
}

func (u *userProtoMapper) ToProtoResponseUserAll(status string, message string) *userpb.ApiResponseUserAll {
	return &userpb.ApiResponseUserAll{
		Status:  status,
		Message: message,
	}
}

func (u *userProtoMapper) ToProtoResponsePaginationUserDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, users []*response.UserResponseDeleteAt) *userpb.ApiResponsePaginationUserDeleteAt {
	return &userpb.ApiResponsePaginationUserDeleteAt{
		Status:     status,
		Message:    message,
		Data:       u.mapResponsesUserDeleteAt(users),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (u *userProtoMapper) ToProtoResponsePaginationUser(pagination *commonpb.PaginationMeta, status string, message string, users []*response.UserResponse) *userpb.ApiResponsePaginationUser {
	return &userpb.ApiResponsePaginationUser{
		Status:     status,
		Message:    message,
		Data:       u.mapResponsesUser(users),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (u *userProtoMapper) mapResponseUserDeleteAt(user *response.UserResponseDeleteAt) *userpb.UserResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if user.DeletedAt != nil {
		deletedAt = wrapperspb.String(*user.DeletedAt)
	}

	return &userpb.UserResponseDeleteAt{
		Id:        int32(user.ID),
		Firstname: user.FirstName,
		Lastname:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func (u *userProtoMapper) mapResponseUser(user *response.UserResponse) *userpb.UserResponse {
	return &userpb.UserResponse{
		Id:        int32(user.ID),
		Firstname: user.FirstName,
		Lastname:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (u *userProtoMapper) mapResponsesUser(users []*response.UserResponse) []*userpb.UserResponse {
	var mappedUsers []*userpb.UserResponse

	for _, user := range users {
		mappedUsers = append(mappedUsers, u.mapResponseUser(user))
	}

	return mappedUsers
}

func (u *userProtoMapper) mapResponseUserDelete(user *response.UserResponseDeleteAt) *userpb.UserResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if user.DeletedAt != nil {
		deletedAt = wrapperspb.String(*user.DeletedAt)
	}

	return &userpb.UserResponseDeleteAt{
		Id:        int32(user.ID),
		Firstname: user.FirstName,
		Lastname:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func (u *userProtoMapper) mapResponsesUserDeleteAt(users []*response.UserResponseDeleteAt) []*userpb.UserResponseDeleteAt {
	var mappedUsers []*userpb.UserResponseDeleteAt

	for _, user := range users {
		mappedUsers = append(mappedUsers, u.mapResponseUserDelete(user))
	}

	return mappedUsers
}

func mapPaginationMeta(s *commonpb.PaginationMeta) *commonpb.PaginationMeta {
	return &commonpb.PaginationMeta{
		CurrentPage:  int32(s.CurrentPage),
		PageSize:     int32(s.PageSize),
		TotalPages:   int32(s.TotalPages),
		TotalRecords: int32(s.TotalRecords),
	}
}
