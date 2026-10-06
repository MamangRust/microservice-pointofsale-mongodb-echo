package repository

import (
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	useradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	MerchantQuery           MerchantQueryRepository
	MerchantCommand         MerchantCommandRepository
	MerchantDocumentQuery   MerchantDocumentQueryRepository
	MerchantDocumentCommand MerchantDocumentCommandRepository
	UserQuery               *useradapter.Repository
}

// GuardOptions collects the dependency guards for each remote dependency.
type GuardOptions struct {
	User []adapter.GuardOption
}

func NewRepositories(db *mongo.Database, userClient pbuser.UserServiceClient, guards ...GuardOptions) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		MerchantQuery:           NewMerchantQueryRepository(db),
		MerchantCommand:         NewMerchantCommandRepository(db),
		MerchantDocumentQuery:   NewMerchantDocumentQueryRepository(db),
		MerchantDocumentCommand: NewMerchantDocumentCommandRepository(db),
		UserQuery:               useradapter.New(userClient, g.User...),
	}
}
