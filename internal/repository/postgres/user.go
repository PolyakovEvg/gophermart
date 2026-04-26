package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	ID           string
	Login        string
	PasswordHash string
}

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, login, passwordHash string) (string, error) {
	var userID string

	err := r.DB.QueryRowContext(ctx,
		"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
		login, passwordHash).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
			strings.Contains(err.Error(), "users_login_key") {
			return "", ErrUserExists
		}
		return "", err
	}
	return userID, nil
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	u := &User{}
	err := r.DB.QueryRowContext(ctx,
		"SELECT id, login, password_hash FROM users WHERE login = $1",
		login).Scan(&u.ID, &u.Login, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}
