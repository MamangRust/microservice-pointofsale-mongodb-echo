package repository

import (
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	roleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/role"
	userroleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user_role"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type GuardOptions struct {
	UserRole []adapter.GuardOption
	Role     []adapter.GuardOption
}

type Repositories struct {
	UserCommand UserCommandRepository
	UserQuery   UserQueryRepository
	Role        RoleQueryRepository
	UserRole    UserRoleRepository
}

// Compile-time assertions: the shared adapters must satisfy the user service
// repository contracts so the wiring can inject them directly.
var (
	_ RoleQueryRepository = (*roleadapter.Repository)(nil)
	_ UserRoleRepository  = (*userroleadapter.Repository)(nil)
)

type Deps struct {
	Db              *mongo.Database
	RoleQueryClient pbrole.RoleServiceClient
	UserRoleClient  pbuserrole.UserRoleServiceClient
	Guard           GuardOptions
}

func NewRepositories(deps *Deps) *Repositories {
	return &Repositories{
		UserCommand: NewUserCommandRepository(deps.Db),
		UserQuery:   NewUserQueryRepository(deps.Db),
		Role:        roleadapter.New(deps.RoleQueryClient, deps.Guard.Role...),
		UserRole:    userroleadapter.New(deps.UserRoleClient, deps.Guard.UserRole...),
	}
}
