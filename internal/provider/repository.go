package provider

import "go-musthave-diploma-tpl/internal/repository/postgres"

func NewUserRepository(store *postgres.DBStorage) *postgres.UserRepository {
	return postgres.NewUserRepository(store.DB)
}
