package user

import (
	"RealTime/internal/auth"
	"RealTime/internal/config"
	userservice "RealTime/internal/core/service"
	"RealTime/internal/log"
	"RealTime/internal/transport/contracts"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type UserAPI struct {
	UserService      contracts.UserServiceProvider
	MessagePublisher contracts.Publisher
	Config           *config.Config // <-- 4. FIXED Config type
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

func NewUserAPI(deps *contracts.AppDependencies) *UserAPI {
	return &UserAPI{
		UserService:      deps.UserService,
		MessagePublisher: deps.Publisher,
		Config:           deps.Config,
	}
}

func (a *UserAPI) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	u, err := a.UserService.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		// --- 5. FIXED: Handle specific domain errors ---
		if errors.Is(err, userservice.ErrUsernameTaken) {
			http.Error(w, "Username is already taken", http.StatusConflict) // 409 Conflict
			return
		}

		// Fallback for all other errors
		log.Logger.Error("Failed to register user", zap.Error(err), zap.String("username", req.Username))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Publish side-effect (correctly done)
	eventData, _ := json.Marshal(map[string]string{
		"type":    "USER_REGISTERED",
		"user_id": u.ID.String(),
	})
	if err := a.MessagePublisher.Publish("user_events", eventData); err != nil {
		log.Logger.Error("Failed to publish user registered event", zap.Error(err), zap.String("user_id", u.ID.String()))
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"id":       u.ID.String(),
		"username": u.Username,
	})
}

func (a *UserAPI) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	u, err := a.UserService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, userservice.ErrInvalidCredentials) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Fallback for all other errors
		log.Logger.Error("Failed to login user", zap.Error(err), zap.String("username", req.Username))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateJWT(
		u.ID.String(),
		u.Username,
		a.Config.JWTSecret,
		a.Config.TokenTimeout,
	)
	if err != nil {
		log.Logger.Error("Failed to generate JWT", zap.Error(err), zap.String("user_id", u.ID.String()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, authResponse{
		Token:    token,
		UserID:   u.ID.String(),
		UserName: u.Username,
	})
}

// respondJSON is a reusable helper
func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// Use the imported logger
		log.Logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}
