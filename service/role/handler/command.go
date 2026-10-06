package handler

import (
	"context"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/role"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	role_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/role_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *roleQueryHandleGrpc) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.ApiResponseRole, error) {
	reqService := &requests.CreateRoleRequest{Name: req.GetName()}
	if err := reqService.Validate(); err != nil {
		return nil, role_errors.ErrGrpcValidateCreateRole
	}
	role, err := s.roleCommandService.CreateRole(ctx, reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRole{Status: "success", Message: "Successfully created role", Data: mapRoleModel(role)}, nil
}

func (s *roleQueryHandleGrpc) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.ApiResponseRole, error) {
	roleID := int(req.GetId())
	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}
	reqService := &requests.UpdateRoleRequest{ID: &roleID, Name: req.GetName()}
	if err := reqService.Validate(); err != nil {
		return nil, role_errors.ErrGrpcValidateUpdateRole
	}
	role, err := s.roleCommandService.UpdateRole(ctx, reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRole{Status: "success", Message: "Successfully updated role", Data: mapRoleModel(role)}, nil
}

func (s *roleQueryHandleGrpc) TrashedRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRole, error) {
	roleID := int(req.GetRoleId())
	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}
	role, err := s.roleCommandService.TrashedRole(ctx, roleID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRole{Status: "success", Message: "Successfully trashed role", Data: mapRoleModel(role)}, nil
}

func (s *roleQueryHandleGrpc) RestoreRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRole, error) {
	roleID := int(req.GetRoleId())
	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}
	role, err := s.roleCommandService.RestoreRole(ctx, roleID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRole{Status: "success", Message: "Successfully restored role", Data: mapRoleModel(role)}, nil
}

func (s *roleQueryHandleGrpc) DeleteRolePermanent(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRoleDelete, error) {
	id := int(req.GetRoleId())
	if id <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}
	_, err := s.roleCommandService.DeleteRolePermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRoleDelete{Status: "success", Message: "Successfully deleted role permanently"}, nil
}

func (s *roleQueryHandleGrpc) RestoreAllRole(ctx context.Context, req *emptypb.Empty) (*pb.ApiResponseRoleAll, error) {
	_, err := s.roleCommandService.RestoreAllRole(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRoleAll{Status: "success", Message: "Successfully restored all roles"}, nil
}

func (s *roleQueryHandleGrpc) DeleteAllRolePermanent(ctx context.Context, req *emptypb.Empty) (*pb.ApiResponseRoleAll, error) {
	_, err := s.roleCommandService.DeleteAllRolePermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseRoleAll{Status: "success", Message: "Successfully deleted all roles permanently"}, nil
}
