package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormPermissionRepository struct {
	db      *gorm.DB
	rdb     *redis.Client
	permTTL time.Duration
}

func NewGormPermissionRepository(db *gorm.DB) domain.PermissionRepository {
	return &GormPermissionRepository{db: db}
}

func NewGormPermissionRepositoryWithCache(db *gorm.DB, rdb *redis.Client, ttl time.Duration) domain.PermissionRepository {
	return &GormPermissionRepository{db: db, rdb: rdb, permTTL: ttl}
}

func userPermKey(userID string) string     { return fmt.Sprintf("auth:perm:user:%s", userID) }
func clientPermKey(clientID string) string { return fmt.Sprintf("auth:perm:client:%s", clientID) }

type grantsPayload struct {
	Roles  []string `json:"roles"`
	Scopes []string `json:"scopes"`
}

// ListRoles implements domain.PermissionRepository.
func (g *GormPermissionRepository) ListRoles() ([]domain.Role, error) {
	var roles []domain.Role
	if err := g.db.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GrantUserScope implements domain.PermissionRepository.
func (g *GormPermissionRepository) GrantUserScope(userID, scopeID, grantedBy string, expiresAt *time.Time) error {
	rec := model.UserScope{
		UserID:    userID,
		ScopeID:   scopeID,
		GrantedBy: grantedBy,
		ExpiresAt: expiresAt,
	}
	if err := g.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "scope_id"}},
		DoUpdates: clause.Assignments(map[string]any{"expires_at": expiresAt, "granted_by": grantedBy}),
	}).Create(&rec).Error; err != nil {
		return err
	}
	return g.InvalidateUser(userID)
}

// RevokeUserScope implements domain.PermissionRepository.
func (g *GormPermissionRepository) RevokeUserScope(userID, scopeID string) error {
	if err := g.db.Delete(&model.UserScope{}, "user_id = ? AND scope_id = ?", userID, scopeID).Error; err != nil {
		return err
	}
	return g.InvalidateUser(userID)
}

// ListUserRoles implements domain.PermissionRepository.
func (g *GormPermissionRepository) ListUserRoles(userID string) ([]string, error) {
	var rows []struct{ Key string }
	q := g.db.Table("roles r").
		Select("r.key").
		Joins("JOIN user_roles ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Group("r.key")
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, v := range rows {
		out = append(out, v.Key)
	}
	return out, nil
}

func (g *GormPermissionRepository) ListUserScopesEffective(userID string, now time.Time) (roles []string, scopes []string, err error) {
	if g.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if b, errGet := g.rdb.Get(ctx, userPermKey(userID)).Bytes(); errGet == nil {
			var p grantsPayload
			if errUnmarshal := json.Unmarshal(b, &p); errUnmarshal == nil {
				return p.Roles, p.Scopes, nil
			}
		}
	}

	var roleScopes []struct{ Key string }
	if err = g.db.Table("scopes s").
		Select("s.key").
		Joins("JOIN role_scopes rs ON rs.scope_id = s.id").
		Joins("JOIN user_roles ur ON ur.role_id = rs.role_id").
		Where("ur.user_id = ?", userID).
		Group("s.key").
		Find(&roleScopes).Error; err != nil {
		return nil, nil, err
	}

	var directScopes []struct{ Key string }
	if err = g.db.Table("scopes s").
		Select("s.key").
		Joins("JOIN user_scopes us ON us.scope_id = s.id").
		Where("us.user_id = ?", userID).
		Where(g.db.Where("us.expires_at IS NULL").Or("us.expires_at > ?", now)).
		Group("s.key").
		Find(&directScopes).Error; err != nil {
		return nil, nil, err
	}

	roles, err = g.ListUserRoles(userID)
	if err != nil {
		return nil, nil, err
	}

	set := make(map[string]struct{}, len(roleScopes)+len(directScopes))
	for _, s := range roleScopes {
		set[s.Key] = struct{}{}
	}
	for _, s := range directScopes {
		set[s.Key] = struct{}{}
	}
	eff := make([]string, 0, len(set))
	for k := range set {
		eff = append(eff, k)
	}

	if g.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = g.setUserGrantsCache(ctx, userID, roles, eff)
	}

	return roles, eff, nil
}

// AddRolesToUser implements domain.PermissionRepository.
func (g *GormPermissionRepository) AddRolesToUser(userID string, roleIDs []string) error {
	ur := make([]model.UserRole, len(roleIDs))
	for i, roleID := range roleIDs {
		ur[i] = model.UserRole{UserID: userID, RoleID: roleID}
	}
	if err := g.db.Create(&ur).Error; err != nil {
		return err
	}
	return g.InvalidateUser(userID)
}

// AddScopesToClient implements domain.PermissionRepository.
func (g *GormPermissionRepository) AddScopesToClient(clientID string, scopeIDs []string) error {
	sc := make([]model.ClientScope, len(scopeIDs))
	for i, scopeID := range scopeIDs {
		sc[i] = model.ClientScope{ClientID: clientID, ScopeID: scopeID}
	}
	if err := g.db.Create(&sc).Error; err != nil {
		return err
	}
	return g.InvalidateClient(clientID)
}

// AddScopesToRole implements domain.PermissionRepository.
func (g *GormPermissionRepository) AddScopesToRole(roleID string, scopeIDs []string) error {
	sr := make([]model.RoleScope, len(scopeIDs))
	for i, scopeID := range scopeIDs {
		sr[i] = model.RoleScope{RoleID: roleID, ScopeID: scopeID}
	}
	return g.db.Create(&sr).Error
}

// CreateRole implements domain.PermissionRepository.
func (g *GormPermissionRepository) CreateRole(key string, desc string) (domain.Role, error) {
	r := domain.Role{Key: key, Desc: desc}
	if err := g.db.Create(&r).Error; err != nil {
		return domain.Role{}, err
	}
	return r, nil
}

// CreateScope implements domain.PermissionRepository.
func (g *GormPermissionRepository) CreateScope(key string, desc string) (domain.Scope, error) {
	s := domain.Scope{Key: key, Desc: desc}
	if err := g.db.Create(&s).Error; err != nil {
		return domain.Scope{}, err
	}
	return s, nil
}

// InvalidateClient implements domain.PermissionRepository.
func (g *GormPermissionRepository) InvalidateClient(clientID string) error {
	if g.rdb == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return g.rdb.Del(ctx, clientPermKey(clientID)).Err()
}

// InvalidateUser implements domain.PermissionRepository.
func (g *GormPermissionRepository) InvalidateUser(userID string) error {
	if g.rdb == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return g.rdb.Del(ctx, userPermKey(userID)).Err()
}

// ListClientScopes implements domain.PermissionRepository.
func (g *GormPermissionRepository) ListClientScopes(clientID string) ([]string, error) {
	if g.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if b, errGet := g.rdb.Get(ctx, clientPermKey(clientID)).Bytes(); errGet == nil {
			var p grantsPayload
			if errUnmarshal := json.Unmarshal(b, &p); errUnmarshal == nil {
				return p.Scopes, nil
			}
		}
	}

	var rows []struct{ Key string }
	q := g.db.Table("scopes s").
		Select("s.key").
		Joins("JOIN client_scopes cs ON cs.scope_id = s.id").
		Where("cs.client_id = ?", clientID).
		Group("s.key")
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Key)
	}

	if g.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		p := grantsPayload{Roles: nil, Scopes: out}
		if b, _ := json.Marshal(p); b != nil {
			_ = g.rdb.Set(ctx, clientPermKey(clientID), b, g.permTTL).Err()
		}
	}

	return out, nil
}

// ListScopes implements domain.PermissionRepository.
func (g *GormPermissionRepository) ListScopes() ([]domain.Scope, error) {
	var rows []model.Scope
	if err := g.db.Order("key").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Scope, 0, len(rows))
	for _, s := range rows {
		out = append(out, domain.Scope{ID: s.ID, Key: s.Key, Desc: s.Description})
	}
	return out, nil
}

// ListUserScopes implements domain.PermissionRepository.
func (g *GormPermissionRepository) ListUserScopes(userID string) ([]string, error) {
	var rows []struct{ Key string }
	q := g.db.Table("scopes s").
		Select("s.key").
		Joins("JOIN role_scopes rs ON rs.scope_id = s.id").
		Joins("JOIN user_roles ur ON ur.role_id = rs.role_id").
		Where("ur.user_id = ?", userID).
		Group("s.key")
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Key)
	}
	return out, nil
}

func (g *GormPermissionRepository) setUserGrantsCache(ctx context.Context, userID string, roles, scopes []string) error {
	if g.rdb == nil {
		return nil
	}
	p := grantsPayload{Roles: roles, Scopes: scopes}
	b, _ := json.Marshal(p)
	return g.rdb.Set(ctx, userPermKey(userID), b, g.permTTL).Err()
}
