package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/cache"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/usecase"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	Signup       *usecase.SignupUseCase
	Login        *usecase.LoginUseCase
	Validate     *validator.Validate
	TokenService domain.TokenService
	Cache        *cache.RedisClient
}

// SignUpRequest represents the payload for user registration
type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"123456"`
}

// LoginRequest represents the payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"123456"`
}

// AuthResponse represents the response containing JWT and user info
type AuthResponse struct {
	Token      string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ID         string `json:"id" example:"123"`
	Name       string `json:"name" example:"John Doe"`
	Email      string `json:"email" example:"user@example.com"`
	Role       string `json:"role" example:"user"`
	Expiration int64  `json:"expiration" example:"1617181723"`
}

// UserResponse represents basic user information
type UserResponse struct {
	ID    string `json:"id" example:"123"`
	Email string `json:"email" example:"user@example.com"`
}

// SignUpHandler godoc
// @Summary Register a new user
// @Description Creates a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param input body SignUpRequest true "User registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Router /auth/signup [post]
func (h *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest); return
	}

	if err := h.Validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest); return
	}

	user, err := h.Signup.Execute(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Error signing up: "+err.Error(), http.StatusInternalServerError); return
	}

	tokenStr, err := h.TokenService.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Error generating token: "+err.Error(), http.StatusInternalServerError); return
	}

	resp := AuthResponse{
		Token:      tokenStr,
		ID:         user.ID,
		Email:      user.Email,
		Expiration: time.Now().Add(h.TokenService.AccessTokenExpiration(tokenStr)).Unix(),
	}

	// (Opcional) cache de profile por ID, TTL curto
	_ = h.Cache.SetJSON("profile:"+user.ID, UserResponse{ID: user.ID, Email: user.Email}, 5*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// LoginHandler godoc
// @Summary Authenticate a user
// @Description Logs in a user with email and password, returning a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "User login credentials"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest); return
	}
	if err := h.Validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest); return
	}

	user, err := h.Login.Execute(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Error logging in: "+err.Error(), http.StatusUnauthorized); return
	}

	tokenStr, err := h.TokenService.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Error generating token: "+err.Error(), http.StatusInternalServerError); return
	}

	resp := AuthResponse{
		Token:      tokenStr,
		ID:         user.ID,
		Email:      user.Email,
		Expiration: time.Now().Add(h.TokenService.AccessTokenExpiration(tokenStr)).Unix(),
	}

	_ = h.Cache.SetJSON("profile:"+user.ID, UserResponse{ID: user.ID, Email: user.Email}, 5*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// MeHandler godoc
// @Summary Get authenticated user info
// @Description Returns the authenticated user's ID and email from the provided JWT token
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized); return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := h.TokenService.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "Invalid or expired token: "+err.Error(), http.StatusUnauthorized); return
	}

	var cached UserResponse
	if err := h.Cache.GetJSON("profile:"+claims.ID, &cached); err == nil && cached.ID != "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cached); return
	}

	resp := UserResponse{ID: claims.ID, Email: claims.Email}
	_ = h.Cache.SetJSON("profile:"+claims.ID, resp, 5*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
