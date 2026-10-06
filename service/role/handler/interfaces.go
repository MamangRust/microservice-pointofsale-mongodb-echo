package handler

import (
	pb "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
)

type RoleQueryHandleGrpc interface {
	pb.RoleServiceServer
}

type UserRoleHandleGrpc interface {
	pbuserrole.UserRoleServiceServer
}
