package service

import (
	"context"
	"errors"
	"fmt"
	"go-musthave-diploma-tpl/internal/repository/postgres"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
)

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrLoginPasswordRequired = errors.New("login and password required")
	ErrLoginPasswordEqual    = errors.New("login and password should not be equal")
	ErrUserExists            = errors.New("user already exists")
)

var (
	ErrPasswordTooShort = fmt.Errorf("password too short: need at least %d characters", MinPasswordLength)
	ErrPasswordTooLong  = fmt.Errorf("password too long: maximum %d characters", MaxPasswordLength)
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string, logger *zap.Logger) (string, error)
	GetUserByLogin(ctx context.Context, login string, logger *zap.Logger) (*postgres.User, error)
}

type AuthService struct {
	userRepo UserRepository
	secret   string
	logger   *zap.Logger
}

func NewAuthService(userRepo UserRepository, secret string, logger *zap.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		secret:   secret,
		logger:   logger,
	}
}

func checkLogin(login, password string) error {
	if login == "" || password == "" {
		return ErrLoginPasswordRequired
	}
	if login == password {
		return ErrLoginPasswordEqual
	}
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}
	return nil
}

func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	if err := checkLogin(login, password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := s.userRepo.CreateUser(ctx, login, string(hash), s.logger)
	if err != nil {
		if errors.Is(err, postgres.ErrUserExists) {
			return "", ErrUserExists
		}
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := GenerateToken(userID, s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	if err := checkLogin(login, password); err != nil {
		return "", err
	}

	user, err := s.userRepo.GetUserByLogin(ctx, login, s.logger)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			s.logger.Info("user not found", zap.String("login", login))
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := GenerateToken(user.ID, s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}
