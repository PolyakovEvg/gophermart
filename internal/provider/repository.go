 package provider

import (
	"go-musthave-diploma-tpl/internal/repository/postgres"
)

func NewUserRepository(store *postgres.DBStorage) *postgres.UserRepository {
	return postgres.NewUserRepository(store.DB)
}

func NewOrderRepository(store *postgres.DBStorage) *postgres.OrderRepository {
	return postgres.NewOrderRepository(store.DB)
}

func NewWithdrawalRepository(store *postgres.DBStorage) *postgres.WithdrawalRepository {
	return postgres.NewWithdrawalRepository(store.DB)
}
