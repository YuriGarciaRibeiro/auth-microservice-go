package domain

import "time"

type Scope struct {
	ID   string
	Key  string
	Desc string
}

type Role struct {
	ID   string
	Key  string
	Desc string
}

type PermissionRepository interface {
	// Catálogo/Admin
	CreateScope(key, desc string) (Scope, error)
	ListScopes() ([]Scope, error)
	CreateRole(key, desc string) (Role, error)
	AddScopesToRole(roleID string, scopeIDs []string) error
	AddRolesToUser(userID string, roleIDs []string) error
	AddScopesToClient(clientID string, scopeIDs []string) error
	ListRoles() ([]Role, error)

	// Consulta (para emissão/checagem)
	ListUserRoles(userID string) ([]string, error)
	ListUserScopesEffective(userID string, now time.Time) (roles []string, scopes []string, err error)
	ListClientScopes(clientID string) ([]string, error)

	// Exceções (scopes diretos ao usuário)
	GrantUserScope(userID, scopeID, grantedBy string, expiresAt *time.Time) error
	RevokeUserScope(userID, scopeID string) error

	// Cache invalidation hooks
	InvalidateUser(userID string) error
	InvalidateClient(clientID string) error
	ListUserScopes(userID string) ([]string, error)
}
