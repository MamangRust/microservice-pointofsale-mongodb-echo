package protomapper

import (
	commonpb "github.com/MamangRust/microservice-point-of-sale-pb/common"
	rolepb "github.com/MamangRust/microservice-point-of-sale-pb/role"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type roleProtoMapper struct {
}

func NewRoleProtoMapper() *roleProtoMapper {
	return &roleProtoMapper{}
}

func (s *roleProtoMapper) ToProtoResponseRoleAll(status string, message string) *rolepb.ApiResponseRoleAll {
	return &rolepb.ApiResponseRoleAll{
		Status:  status,
		Message: message,
	}
}

func (s *roleProtoMapper) ToProtoResponseRoleDelete(status string, message string) *rolepb.ApiResponseRoleDelete {
	return &rolepb.ApiResponseRoleDelete{
		Status:  status,
		Message: message,
	}
}

func (s *roleProtoMapper) ToProtoResponseRole(status string, message string, pbResponse *response.RoleResponse) *rolepb.ApiResponseRole {
	return &rolepb.ApiResponseRole{
		Status:  status,
		Message: message,
		Data:    s.mapResponseRole(pbResponse),
	}
}

func (s *roleProtoMapper) ToProtoResponsesRole(status string, message string, pbResponse []*response.RoleResponse) *rolepb.ApiResponsesRole {
	return &rolepb.ApiResponsesRole{
		Status:  status,
		Message: message,
		Data:    s.mapResponsesRole(pbResponse),
	}
}

func (s *roleProtoMapper) ToProtoResponsePaginationRole(pagination *commonpb.PaginationMeta, status string, message string, pbResponse []*response.RoleResponse) *rolepb.ApiResponsePaginationRole {
	return &rolepb.ApiResponsePaginationRole{
		Status:     status,
		Message:    message,
		Data:       s.mapResponsesRole(pbResponse),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (s *roleProtoMapper) ToProtoResponsePaginationRoleDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, pbResponse []*response.RoleResponseDeleteAt) *rolepb.ApiResponsePaginationRoleDeleteAt {
	return &rolepb.ApiResponsePaginationRoleDeleteAt{
		Status:     status,
		Message:    message,
		Data:       s.mapResponsesRoleDeleteAt(pbResponse),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (s *roleProtoMapper) mapResponseRole(role *response.RoleResponse) *rolepb.RoleResponse {
	return &rolepb.RoleResponse{
		Id:        int32(role.ID),
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}

func (s *roleProtoMapper) mapResponsesRole(roles []*response.RoleResponse) []*rolepb.RoleResponse {
	var responseRoles []*rolepb.RoleResponse

	for _, role := range roles {
		responseRoles = append(responseRoles, s.mapResponseRole(role))
	}

	return responseRoles
}

func (s *roleProtoMapper) mapResponseRoleDeleteAt(role *response.RoleResponseDeleteAt) *rolepb.RoleResponseDeleteAt {
	return &rolepb.RoleResponseDeleteAt{
		Id:        int32(role.ID),
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
		DeletedAt: role.DeletedAt,
	}
}

func (s *roleProtoMapper) mapResponsesRoleDeleteAt(roles []*response.RoleResponseDeleteAt) []*rolepb.RoleResponseDeleteAt {
	var responseRoles []*rolepb.RoleResponseDeleteAt

	for _, role := range roles {
		responseRoles = append(responseRoles, s.mapResponseRoleDeleteAt(role))
	}

	return responseRoles
}
