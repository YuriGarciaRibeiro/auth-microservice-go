package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	apierrors "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/errors"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/cache"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/usecase"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	Signup               *usecase.SignupUseCase
	Login                *usecase.LoginUseCase
	Validate             *validator.Validate
	TokenService         domain.TokenService
	Cache                *cache.RedisClient
	PermissionRepository domain.PermissionRepository
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"123456"`
}

// AuthResponse is returned on successful authentication.
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	AccessExp    time.Time `json:"access_exp"`
	RefreshExp   time.Time `json:"refresh_exp"`
}

// UserResponse represents basic user information
type UserResponse struct {
	ID    string `json:"id" example:"123"`
	Email string `json:"email" example:"user@example.com"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	AccessExp    time.Time `json:"access_exp"`
	RefreshExp   time.Time `json:"refresh_exp"`
}

type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"123456"`
}

type LogoutRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type IntrospectRequest struct {
	Token string `json:"token" validate:"required"`
}

type IntrospectResponse struct {
	Active      bool     `json:"active"`
	SubjectType string   `json:"subject_type,omitempty"`
	Sub         string   `json:"sub,omitempty"`
	Email       string   `json:"email,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Scope       []string `json:"scope,omitempty"`
	Aud         []string `json:"aud,omitempty"`
}

// SignUpHandler godoc
// @Summary Register a new user
// @Description Creates a new user account and returns an access+refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param input body SignUpRequest true "User registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/signup [post]
func (h *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	// 1) Decode and validate payload
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}

	// 2) Create user (domain-level)
	user, err := h.Signup.Execute(req.Email, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			apierrors.Conflict(w, "User with this email already exists")
			return
		}
		apierrors.InternalError(w, "Failed to create user")
		return
	}

	roles, scopes, err := h.PermissionRepository.ListUserScopesEffective(user.ID, time.Now())
	if err != nil {
		apierrors.InternalError(w, "Failed to fetch user permissions")
		return
	}

	// 3) Build principal and issue tokens (access + refresh)
	principal := domain.Principal{
		Type:     domain.PrincipalUser,
		ID:       user.ID,
		Email:    user.Email,
		Roles:    roles,
		Scopes:   scopes,
		Audience: nil,
	}

	pair, err := h.TokenService.IssuePair(principal)
	if err != nil {
		apierrors.InternalError(w, "Failed to issue authentication tokens")
		return
	}

	if h.Cache != nil {
		_ = h.Cache.SetJSON("profile:"+user.ID, UserResponse{ID: user.ID, Email: user.Email}, 5*time.Minute)
	}

	// 5) Respond
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		AccessExp:    pair.AccessExp,
		RefreshExp:   pair.RefreshExp,
	})
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
	// 1) Decode and validate input
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}

	user, err := h.Login.Execute(req.Email, req.Password)
	if err != nil || user.ID == "" {
		apierrors.Unauthorized(w, "Invalid email or password")
		return
	}

	roles, scopes, err := h.PermissionRepository.ListUserScopesEffective(user.ID, time.Now())
	if err != nil {
		apierrors.InternalError(w, "Failed to fetch user permissions")
		return
	}

	principal := domain.Principal{
		Type:     domain.PrincipalUser,
		ID:       user.ID,
		Email:    req.Email,
		Roles:    roles,
		Scopes:   scopes,
		Audience: nil,
	}

	pair, err := h.TokenService.IssuePair(principal)
	if err != nil {
		apierrors.InternalError(w, "Failed to issue authentication tokens")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		AccessExp:    pair.AccessExp,
		RefreshExp:   pair.RefreshExp,
	})
}

// RefreshHandler godoc
// @Summary Rotate tokens using a valid refresh token
// @Description Exchanges a valid refresh token for a new access+refresh pair
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RefreshRequest true "Refresh token payload"
// @Success 200 {object} RefreshResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}

	pair, err := h.TokenService.Rotate(req.RefreshToken)
	if err != nil {
		apierrors.Unauthorized(w, "Invalid or expired refresh token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(RefreshResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		AccessExp:    pair.AccessExp,
		RefreshExp:   pair.RefreshExp,
	})
}

// LogoutHandler godoc
// @Summary Logout and revoke tokens
// @Description Revokes the provided tokens: access is blacklisted; refresh is removed from Redis.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LogoutRequest true "Tokens to revoke"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}

	access := req.AccessToken
	if access == "" {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			access = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	if err := h.TokenService.RevokePair(access, req.RefreshToken); err != nil {
		apierrors.InternalError(w, "Failed to revoke tokens")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// IntrospectHandler godoc
// @Summary Introspect an access token
// @Description Validates an access token and returns whether it's active along with principal info
// @Tags auth
// @Accept json
// @Produce json
// @Param input body IntrospectRequest true "Token to introspect"
// @Success 200 {object} IntrospectResponse
// @Failure 400 {object} map[string]string
// @Router /auth/introspect [post]
func (h *AuthHandler) IntrospectHandler(w http.ResponseWriter, r *http.Request) {
	var req IntrospectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}

	active, claims, err := h.TokenService.Introspect(req.Token)
	if err != nil {
		// Operational error on our side (e.g., Redis down)
		apierrors.InternalError(w, "Token introspection failed")
		return
	}

	resp := IntrospectResponse{Active: active}
	if active && claims != nil {
		resp.SubjectType = string(claims.SubjectType)
		resp.Sub = claims.SubjectID
		resp.Email = claims.Email
		resp.Roles = claims.Roles
		resp.Scope = claims.Scopes
		resp.Aud = claims.Audience
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
