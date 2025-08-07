package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/usecase"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	Signup       *usecase.SignupUseCase
	Login        *usecase.LoginUseCase
	Validate     *validator.Validate
	TokenService domain.TokenService
}

type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type AuthResponse struct {
	Token      string `json:"token"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Expiration int64  `json:"expiration"`
}

func (h *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.Signup.Execute(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Error signing up: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokenStr, err := h.TokenService.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Error generating token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token:      tokenStr,
		ID:         user.ID,
		Email:      user.Email,
		Expiration: time.Now().Add(h.TokenService.AccessTokenExpiration(tokenStr)).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.Login.Execute(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Error logging in: "+err.Error(), http.StatusUnauthorized)
		return
	}

	tokenStr, err := h.TokenService.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Error generating token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token:      tokenStr,
		ID:         user.ID,
		Email:      user.Email,
		Expiration: time.Now().Add(h.TokenService.AccessTokenExpiration(tokenStr)).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

	// Esperado: "Bearer <token>"
	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	tokenStr := authHeader[len(prefix):]

	// Validar token e extrair claims
	claims, err := h.TokenService.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "Invalid or expired token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"id":    claims.ID,
		"email": claims.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

