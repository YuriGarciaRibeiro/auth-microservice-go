package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	apierrors "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/errors"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/usecase"
	"github.com/go-playground/validator/v10"
)

type ClientTokenRequest struct {
	ClientID string   `json:"client_id" validate:"required"`
	Secret   string   `json:"client_secret" validate:"required"`
	Scopes   []string `json:"scopes"`
	Audience []string `json:"audience"`
}

type ClientTokenResponse struct {
	AccessToken string    `json:"access_token"`
	AccessExp   time.Time `json:"access_exp"`
}

type ClientTokenHandler struct {
	Validate             *validator.Validate
	UC                   *usecase.ClientCredentialsUseCase
	TokenService         domain.TokenService
	PermissionRepository domain.PermissionRepository
}

// @Summary      Client Token
// @Description  Issue access token for client credentials grant
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ClientTokenRequest true "Client credentials payload"
// @Success      200 {object} ClientTokenResponse
// @Failure      400 {string} string "Invalid payload"
// @Failure      422 {string} string "Validation failed"
// @Failure      401 {string} string "Invalid client credentials"
// @Failure      500 {string} string "Internal server error"
// @Router       /auth/token [post]
func (h *ClientTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req ClientTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}

	scopes, err := h.PermissionRepository.ListClientScopes(req.ClientID)
	if err != nil {
		apierrors.Unauthorized(w, "Invalid client credentials")
		return
	}

	principal, err := h.UC.Execute(usecase.ClientCredentialsInput{
		ClientID: req.ClientID,
		Secret:   req.Secret,
		Scopes:   scopes,
		Audience: req.Audience,
	})
	if err != nil {
		apierrors.Unauthorized(w, "Invalid client credentials")
		return
	}

	token, exp, err := h.TokenService.IssueAccessOnly(principal)
	if err != nil {
		apierrors.InternalError(w, "Failed to issue access token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ClientTokenResponse{
		AccessToken: token,
		AccessExp:   exp,
	})
}
