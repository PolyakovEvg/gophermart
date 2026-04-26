package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go-musthave-diploma-tpl/internal/service"
	"net/http"

	"go.uber.org/zap"
)

type AuthServicer interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

type AuthHandler struct {
	authService AuthServicer
	logger      *zap.Logger
}

func NewAuthHandler(authService AuthServicer, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Register(ctx, req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserExists):
			http.Error(w, "login already taken", http.StatusConflict)
		case errors.Is(err, service.ErrLoginPasswordRequired):
			http.Error(w, "login and password required", http.StatusBadRequest)
		case errors.Is(err, service.ErrLoginPasswordEqual):
			http.Error(w, "login and password should not be equal", http.StatusBadRequest)
		case errors.Is(err, service.ErrPasswordTooShort):
			http.Error(w, "password too short", http.StatusBadRequest)
		default:
			h.logger.Error("registration error", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, service.ErrLoginPasswordRequired) {
			http.Error(w, "login and password required", http.StatusBadRequest)
			return
		}
		h.logger.Error("login error", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
