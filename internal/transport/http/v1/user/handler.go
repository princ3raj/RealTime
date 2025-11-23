package user

import (
	"RealTime/internal/auth"
	userservice "RealTime/internal/core/service"
	userdomain "RealTime/internal/domain/user"
	"RealTime/internal/logger"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ServiceProvider defines exactly what we need from the Core
type ServiceProvider interface {
	Register(ctx context.Context, username, password string) (*userdomain.User, error)
	Login(ctx context.Context, username, password string) (*userdomain.User, error)
}

// HandlerConfig extracts only the specific settings this handler needs
type HandlerConfig struct {
	JWTSecret    string
	TokenTimeout time.Duration
}

type API struct {
	svc    ServiceProvider
	config HandlerConfig
}

// NewUserAPI - Notice we don't ask for Publisher here anymore
func NewUserAPI(service ServiceProvider, cfg HandlerConfig) *API {
	return &API{
		svc:    service,
		config: cfg,
	}
}

func (a *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call Service. The Service is now responsible for Publishing events!
	u, err := a.svc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, userservice.ErrUsernameTaken) {
			http.Error(w, "Username is already taken", http.StatusConflict)
			return
		}
		logger.Logger.Error("Failed to register user", zap.Error(err))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Success response
	respondJSON(w, http.StatusCreated, map[string]string{
		"id":       u.ID.String(),
		"username": u.Username,
	})
}

func (a *API) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest // Reusing struct for simplicity, or make LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	u, err := a.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, userservice.ErrInvalidCredentials) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		logger.Logger.Error("Failed to login", zap.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Use the LOCAL config, not global
	token, err := auth.GenerateJWT(
		u.ID.String(),
		u.Username,
		a.config.JWTSecret,
		a.config.TokenTimeout,
	)
	if err != nil {
		logger.Logger.Error("Failed to generate JWT", zap.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Token:    token,
		UserID:   u.ID.String(),
		UserName: u.Username,
	})
}

func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return
	}
}
