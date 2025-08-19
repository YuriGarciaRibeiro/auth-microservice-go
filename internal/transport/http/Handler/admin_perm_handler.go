package handler

import (
	"encoding/json"
	"net/http"
	"time"

	apierrors "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/errors"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/usecase"
	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

type AdminPermHandler struct {
	UC       *usecase.PermAdminUseCase
	Validate *validator.Validate
}

type CreateRoleRequest struct {
	Key  string `json:"key" validate:"required"`
	Desc string `json:"desc" validate:"required"`
}

type CreateScopeRequest struct {
	Key  string `json:"key" validate:"required"`
	Desc string `json:"desc" validate:"required"`
}

type idsBody struct {
	IDs []string `json:"ids" validate:"required,min=1,dive,required"`
}

type grantUserScopeReq struct {
	ScopeID   string     `json:"scope_id" validate:"required"`
	GrantedBy string     `json:"granted_by" validate:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type revokeUserScopeReq struct {
	ScopeID string `json:"scope_id" validate:"required"`
}

// @Summary Create scope
// @Tags    Admin
// @Accept  json
// @Produce json
// @Security BearerAuth
// @Param   request body CreateScopeRequest true "Scope data"
// @Success 201 {object} map[string]string
// @Router  /admin/scopes [post]
func (h *AdminPermHandler) CreateScope(w http.ResponseWriter, r *http.Request) {
	var req CreateScopeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}
	s, err := h.UC.CreateScope(req.Key, req.Desc)
	if err != nil {
		apierrors.Conflict(w, "Scope with this key already exists")
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

// @Summary List scopes
// @Tags    Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]string
// @Router  /admin/scopes [get]
func (h *AdminPermHandler) ListScopes(w http.ResponseWriter, _ *http.Request) {
	list, err := h.UC.ListScopes()
	if err != nil {
		apierrors.InternalError(w, "Internal server error")
		return
	}
	out := make([]map[string]string, 0, len(list))
	for _, s := range list {
		out = append(out, map[string]string{"id": s.ID, "key": s.Key, "desc": s.Desc})
	}
	_ = json.NewEncoder(w).Encode(out)
}

// @Summary Create role
// @Tags    Admin
// @Accept  json
// @Produce json
// @Security BearerAuth
// @Param   request body CreateRoleRequest true "Role data"
// @Success 201 {object} map[string]string
// @Router  /admin/roles [post]
func (h *AdminPermHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}
	role, err := h.UC.CreateRole(req.Key, req.Desc)
	if err != nil {
		apierrors.Conflict(w, "Resource conflict")
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": role.ID, "key": role.Key})
}

// @Summary Attach scopes to role
// @Tags    Admin
// @Accept  json
// @Security BearerAuth
// @Param   roleId path string true "Role ID"
// @Param   request body idsBody true "Scope IDs"
// @Success 204
// @Router  /admin/roles/{roleId}/scopes [post]
func (h *AdminPermHandler) AddScopesToRole(w http.ResponseWriter, r *http.Request) {
	roleID := chiURLParam(r, "roleId")
	var req idsBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}
	if err := h.UC.AddScopesToRole(roleID, req.IDs); err != nil {
		apierrors.Conflict(w, "Resource conflict")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Attach roles to user
// @Tags    Admin
// @Accept  json
// @Security BearerAuth
// @Param   userId path string true "User ID"
// @Param   request body idsBody true "Role IDs"
// @Success 204
// @Router  /admin/users/{userId}/roles [post]
func (h *AdminPermHandler) AddRolesToUser(w http.ResponseWriter, r *http.Request) {
	userID := chiURLParam(r, "userId")
	var req idsBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}
	if err := h.UC.AddRolesToUser(userID, req.IDs); err != nil {
		apierrors.Conflict(w, "Resource conflict")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Attach scopes to client
// @Tags    Admin
// @Accept  json
// @Security BearerAuth
// @Param   clientId path string true "Client ID"
// @Param   request body idsBody true "Scope IDs"
// @Success 204
// @Router  /admin/clients/{clientId}/scopes [post]
func (h *AdminPermHandler) AddScopesToClient(w http.ResponseWriter, r *http.Request) {
	clientID := chiURLParam(r, "clientId")
	var req idsBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}
	if err := h.UC.AddScopesToClient(clientID, req.IDs); err != nil {
		apierrors.Conflict(w, "Resource conflict")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func chiURLParam(r *http.Request, key string) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		return rctx.URLParam(key)
	}
	return ""
}

// @Summary      List roles
// @Description  Returns all roles (id, key, desc)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} map[string]string
// @Router       /admin/roles [get]
func (h *AdminPermHandler) ListRoles(w http.ResponseWriter, _ *http.Request) {
	roles, err := h.UC.ListRoles()
	if err != nil {
		apierrors.InternalError(w, "Internal server error")
		return
	}
	out := make([]map[string]string, 0, len(roles))
	for _, r := range roles {
		out = append(out, map[string]string{"id": r.ID, "key": r.Key, "desc": r.Desc})
	}
	_ = json.NewEncoder(w).Encode(out)
}

// @Summary      List user roles
// @Description  Returns the role keys assigned to the user
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        userId path string true "User ID"
// @Success      200 {array} string
// @Router       /admin/users/{userId}/roles [get]
func (h *AdminPermHandler) ListUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	roles, err := h.UC.ListUserRoles(userID)
	if err != nil {
		apierrors.InternalError(w, "Internal server error")
		return
	}
	_ = json.NewEncoder(w).Encode(roles)
}

// @Summary      List user effective scopes
// @Description  Returns user's effective permissions: roles and final scopes (roles ⊔ direct, non-expired)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        userId path string true "User ID"
// @Success      200 {object} map[string]interface{}
// @Router       /admin/users/{userId}/scopes [get]
func (h *AdminPermHandler) ListUserEffective(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	roles, scopes, err := h.UC.ListUserScopesEffective(userID)
	if err != nil {
		apierrors.InternalError(w, "Internal server error")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"roles": roles, "scopes": scopes})
}

// @Summary      Grant direct scope to user
// @Description  Grants a scope directly to the user (optional expiration). Prefer roles; use direct scopes for exceptions.
// @Tags         Admin
// @Accept       json
// @Security     BearerAuth
// @Param        userId path string true "User ID"
// @Param        request body grantUserScopeReq true "Grant payload"
// @Success      204
// @Failure      400 {string} string "bad request"
// @Failure      422 {string} string "validation failed"
// @Failure      409 {string} string "conflict"
// @Router       /admin/users/{userId}/scopes/grant [post]
func (h *AdminPermHandler) GrantUserScope(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	var req grantUserScopeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}
	if err := h.UC.GrantUserScope(userID, req.ScopeID, req.GrantedBy, req.ExpiresAt); err != nil {
		apierrors.Conflict(w, "Resource conflict")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Revoke direct scope from user
// @Description  Removes a direct scope from the user
// @Tags         Admin
// @Accept       json
// @Security     BearerAuth
// @Param        userId path string true "User ID"
// @Param        request body revokeUserScopeReq true "Revoke payload"
// @Success      204
// @Failure      400 {string} string "bad request"
// @Failure      422 {string} string "validation failed"
// @Failure      409 {string} string "conflict"
// @Router       /admin/users/{userId}/scopes/revoke [post]
func (h *AdminPermHandler) RevokeUserScope(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	var req revokeUserScopeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.BadRequest(w, "Invalid JSON payload")
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		apierrors.ValidationError(w, "Validation failed", err.Error())
		return
	}
	if err := h.UC.RevokeUserScope(userID, req.ScopeID); err != nil {
		apierrors.Conflict(w, "Resource conflict")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      List client scopes
// @Description  Returns scopes attached directly to the client (table client_scopes)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        clientId path string true "Client ID"
// @Success      200 {array} string
// @Router       /admin/clients/{clientId}/scopes [get]
func (h *AdminPermHandler) ListClientScopes(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientId")
	scopes, err := h.UC.ListClientScopes(clientID)
	if err != nil {
		apierrors.InternalError(w, "Internal server error")
		return
	}
	_ = json.NewEncoder(w).Encode(scopes)
}
