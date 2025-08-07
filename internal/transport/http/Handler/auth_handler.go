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
	Signup     *usecase.SignupUseCase
	Login      *usecase.LoginUseCase
	Validate   *validator.Validate
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
	Token     string `json:"token"`
	Id       string `json:"id"`
	Expiration int64  `json:"expiration"`
}

func (h *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and Password are required", http.StatusBadRequest)
		return
	}

	userID, err := h.Signup.Execute(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Error signing up: "+err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.TokenService.GenerateToken(userID)
	if err != nil {
		http.Error(w, "Error generating token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token:     token,
		Id:       userID,
		Expiration: time.Now().Add(h.TokenService.AccessTokenExpiration(token)).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and Password are required", http.StatusBadRequest)
		return
	}

	userID, err := h.Login.Execute(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Error logging in: "+err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.TokenService.GenerateToken(userID)
	if err != nil {
		http.Error(w, "Error generating token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token:     token,
		Id:       userID,
		Expiration: time.Now().Add(h.TokenService.AccessTokenExpiration(token)).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// get all users
func (h *AuthHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := h.Signup.UserRepo.GetAll()
	if err != nil {
		http.Error(w, "Error fetching users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}