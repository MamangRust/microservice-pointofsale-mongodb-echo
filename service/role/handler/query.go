package handler

import (
	"context"
	"math"

	pbutils "github.com/MamangRust/microservice-point-of-sale-pb/common"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/role"
	"github.com/MamangRust/microservice-point-of-sale-role/repository"
	"github.com/MamangRust/microservice-point-of-sale-role/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	role_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/role_errors"
)

type roleQueryHandleGrpc struct {
	pb.UnimplementedRoleServiceServer

	roleQuery service.RoleQueryService

	roleCommandService service.RoleCommandService
}

func NewRoleHandleGrpc(query service.RoleQueryService, command service.RoleCommandService) *roleQueryHandleGrpc {
	return &roleQueryHandleGrpc{
		roleQuery:          query,
		roleCommandService: command,
	}
}

func (s *roleQueryHandleGrpc) FindAllRole(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRole, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{Page: page, PageSize: pageSize, Search: search}
	roles, totalRecords, err := s.roleQuery.FindAll(ctx, &reqService)
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

	return &pb.ApiResponsePaginationRole{
		Status:     "success",
		Message:    "Successfully fetched role records",
		Data:       mapRoleResults(roles),
		Pagination: paginationMeta,
	}, nil
}

func (s *roleQueryHandleGrpc) FindByIdRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRole, error) {
	roleID := int(req.GetRoleId())
	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}
	role, err := s.roleQuery.FindById(ctx, roleID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRole{Status: "success", Message: "Successfully fetched role", Data: mapRoleModel(role)}, nil
}

func (s *roleQueryHandleGrpc) FindByNameRole(ctx context.Context, req *pb.FindByNameRoleRequest) (*pb.ApiResponseRole, error) {
	name := req.GetName()
	if name == "" {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}
	role, err := s.roleQuery.FindByName(ctx, name)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRole{Status: "success", Message: "Successfully fetched role by name", Data: mapRoleModel(role)}, nil
}

func (s *roleQueryHandleGrpc) FindByActive(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRoleDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{Page: page, PageSize: pageSize, Search: search}
	roles, totalRecords, err := s.roleQuery.FindByActiveRole(ctx, &reqService)
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

	return &pb.ApiResponsePaginationRoleDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active roles",
		Data:       mapRoleResultsDeleteAt(roles),
		Pagination: paginationMeta,
	}, nil
}

func (s *roleQueryHandleGrpc) FindByTrashed(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRoleDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{Page: page, PageSize: pageSize, Search: search}
	roles, totalRecords, err := s.roleQuery.FindByTrashedRole(ctx, &reqService)
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

	return &pb.ApiResponsePaginationRoleDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed roles",
		Data:       mapRoleResultsDeleteAt(roles),
		Pagination: paginationMeta,
	}, nil
}

func mapRoleModel(role *models.Role) *pb.RoleResponse {
	if role == nil {
		return nil
	}
	return &pb.RoleResponse{
		Id:        role.RoleID,
		Name:      role.RoleName,
		CreatedAt: convert.FormatTimePtr(role.CreatedAt),
		UpdatedAt: convert.FormatTimePtr(role.UpdatedAt),
	}
}

func mapRoleModels(roles []*models.Role) []*pb.RoleResponse {
	var res []*pb.RoleResponse
	for _, r := range roles {
		res = append(res, mapRoleModel(r))
	}
	return res
}

func mapRoleResults(roles []*repository.RoleResult) []*pb.RoleResponse {
	var res []*pb.RoleResponse
	for _, r := range roles {
		res = append(res, &pb.RoleResponse{
			Id:        r.RoleID,
			Name:      r.RoleName,
			CreatedAt: convert.StrVal(r.CreatedAt),
			UpdatedAt: convert.StrVal(r.UpdatedAt),
		})
	}
	return res
}

func mapRoleResultsDeleteAt(roles []*repository.RoleResult) []*pb.RoleResponseDeleteAt {
	var res []*pb.RoleResponseDeleteAt
	for _, r := range roles {
		res = append(res, &pb.RoleResponseDeleteAt{
			Id:        r.RoleID,
			Name:      r.RoleName,
			CreatedAt: convert.StrVal(r.CreatedAt),
			UpdatedAt: convert.StrVal(r.UpdatedAt),
			DeletedAt: convert.StrVal(r.DeletedAt),
		})
	}
	return res
}
