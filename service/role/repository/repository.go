package repository

import "go.mongodb.org/mongo-driver/v2/mongo"

type Repositories struct {
	RoleCommand     RoleCommandRepository
	RoleQuery       RoleQueryRepository
	UserRoleCommand UserRoleCommandRepository
}

func NewRepositories(db *mongo.Database) *Repositories {
	return &Repositories{
		RoleCommand:     NewRoleCommandRepository(db),
		RoleQuery:       NewRoleQueryRepository(db),
		UserRoleCommand: NewUserRoleCommandRepository(db),
	}
}
