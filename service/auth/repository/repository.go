package repository

import (
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	roleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/role"
	useradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user"
	userroleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user_role"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	UserRole     UserRoleRepository
	Role         RoleRepository
	ResetToken   ResetTokenRepository
}

// GuardOptions lets the wiring layer attach resilience guards to the gRPC
// adapters consumed by this service.
type GuardOptions struct {
	User     []adapter.GuardOption
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

// Compile-time assertions: the shared adapters must satisfy the auth
// repository contracts so the wiring can inject them directly.
var (
	_ UserRepository     = (*useradapter.Repository)(nil)
	_ RoleRepository     = (*roleadapter.Repository)(nil)
	_ UserRoleRepository = (*userroleadapter.Repository)(nil)
)

func NewRepositories(
	db *mongo.Database,
	userClient pbuser.UserServiceClient,
	roleClient pbrole.RoleServiceClient,
	userRoleClient pbuserrole.UserRoleServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		User:         useradapter.New(userClient, g.User...),
		RefreshToken: NewRefreshTokenRepository(db),
		UserRole:     userroleadapter.New(userRoleClient, g.UserRole...),
		Role:         roleadapter.New(roleClient, g.Role...),
		ResetToken:   NewResetTokenRepository(db),
	}
}
