package usecase

import (
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
)

type PermAdminUseCase struct {
	Repo domain.PermissionRepository
	Now  func() time.Time
}

func NewPermAdminUseCase(repo domain.PermissionRepository) *PermAdminUseCase {
	return &PermAdminUseCase{Repo: repo, Now: time.Now}
}

func (uc *PermAdminUseCase) CreateRole(key string, desc string) (domain.Role, error) {
	return uc.Repo.CreateRole(key, desc)
}
func (uc *PermAdminUseCase) CreateScope(key string, desc string) (domain.Scope, error) {
	return uc.Repo.CreateScope(key, desc)
}
func (uc *PermAdminUseCase) AddScopesToRole(roleID string, scopeIDs []string) error {
	return uc.Repo.AddScopesToRole(roleID, scopeIDs)
}
func (uc *PermAdminUseCase) AddRolesToUser(userID string, roleIDs []string) error {
	return uc.Repo.AddRolesToUser(userID, roleIDs)
}
func (uc *PermAdminUseCase) AddScopesToClient(clientID string, scopeIDs []string) error {
	return uc.Repo.AddScopesToClient(clientID, scopeIDs)
}
func (uc *PermAdminUseCase) GrantUserScope(userID string, scopeID string, grantedBy string, expiresAt *time.Time) error {
	return uc.Repo.GrantUserScope(userID, scopeID, grantedBy, expiresAt)
}
func (uc *PermAdminUseCase) RevokeUserScope(userID string, scopeID string) error {
	return uc.Repo.RevokeUserScope(userID, scopeID)
}
func (uc *PermAdminUseCase) InvalidateClient(clientID string) error {
	return uc.Repo.InvalidateClient(clientID)
}
func (uc *PermAdminUseCase) InvalidateUser(userID string) error {
	return uc.Repo.InvalidateUser(userID)
}
func (uc *PermAdminUseCase) ListClientScopes(clientID string) ([]string, error) {
	return uc.Repo.ListClientScopes(clientID)
}
func (uc *PermAdminUseCase) ListUserRoles(userID string) ([]string, error) {
	return uc.Repo.ListUserRoles(userID)
}
func (uc *PermAdminUseCase) ListUserScopesEffective(userID string) (roles []string, scopes []string, err error) {
	return uc.Repo.ListUserScopesEffective(userID, uc.Now())
}
func (uc *PermAdminUseCase) ListRoles() ([]domain.Role, error) {
	return uc.Repo.ListRoles()
}
func (uc *PermAdminUseCase) ListScopes() ([]domain.Scope, error) {
	return uc.Repo.ListScopes()
}
