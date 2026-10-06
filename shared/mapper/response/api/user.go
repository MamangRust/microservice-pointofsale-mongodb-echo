package response_api

import (
	userpb "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type userResponseMapper struct {
}

func NewUserResponseMapper() *userResponseMapper {
	return &userResponseMapper{}
}

func (u *userResponseMapper) ToResponseUser(user *userpb.UserResponse) *response.UserResponse {
	return &response.UserResponse{
		ID:        int(user.Id),
		FirstName: user.Firstname,
		LastName:  user.Lastname,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (u *userResponseMapper) ToResponsesUser(users []*userpb.UserResponse) []*response.UserResponse {
	var mappedUsers []*response.UserResponse

	for _, user := range users {
		mappedUsers = append(mappedUsers, u.ToResponseUser(user))
	}

	return mappedUsers
}

func (u *userResponseMapper) ToResponseUserDeleteAt(user *userpb.UserResponseDeleteAt) *response.UserResponseDeleteAt {
	var deletedAt string
	if user.DeletedAt != nil {
		deletedAt = user.DeletedAt.Value
	}

	return &response.UserResponseDeleteAt{
		ID:        int(user.Id),
		FirstName: user.Firstname,
		LastName:  user.Lastname,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: &deletedAt,
	}
}

func (u *userResponseMapper) ToResponsesUserDeleteAt(users []*userpb.UserResponseDeleteAt) []*response.UserResponseDeleteAt {
	var mappedUsers []*response.UserResponseDeleteAt

	for _, user := range users {
		mappedUsers = append(mappedUsers, u.ToResponseUserDeleteAt(user))
	}

	return mappedUsers
}

func (u *userResponseMapper) ToApiResponseUser(pbResponse *userpb.ApiResponseUser) *response.ApiResponseUser {
	return &response.ApiResponseUser{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    u.ToResponseUser(pbResponse.Data),
	}
}

func (u *userResponseMapper) ToApiResponseUserDeleteAt(pbResponse *userpb.ApiResponseUserDeleteAt) *response.ApiResponseUserDeleteAt {
	return &response.ApiResponseUserDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    u.ToResponseUserDeleteAt(pbResponse.Data),
	}
}

func (u *userResponseMapper) ToApiResponsesUser(pbResponse *userpb.ApiResponsesUser) *response.ApiResponsesUser {
	return &response.ApiResponsesUser{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    u.ToResponsesUser(pbResponse.Data),
	}
}

func (u *userResponseMapper) ToApiResponseUserDelete(pbResponse *userpb.ApiResponseUserDelete) *response.ApiResponseUserDelete {
	return &response.ApiResponseUserDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (u *userResponseMapper) ToApiResponseUserAll(pbResponse *userpb.ApiResponseUserAll) *response.ApiResponseUserAll {
	return &response.ApiResponseUserAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (u *userResponseMapper) ToApiResponsePaginationUserDeleteAt(pbResponse *userpb.ApiResponsePaginationUserDeleteAt) *response.ApiResponsePaginationUserDeleteAt {
	return &response.ApiResponsePaginationUserDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       u.ToResponsesUserDeleteAt(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}

func (u *userResponseMapper) ToApiResponsePaginationUser(pbResponse *userpb.ApiResponsePaginationUser) *response.ApiResponsePaginationUser {
	return &response.ApiResponsePaginationUser{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       u.ToResponsesUser(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}
