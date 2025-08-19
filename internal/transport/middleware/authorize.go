package middleware

import (
	"net/http"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	apierrors "github.com/YuriGarciaRibeiro/auth-microservice-go/internal/errors"
)

// set util for membership checks in O(1)
type stringSet map[string]struct{}

func toSet(ss []string) stringSet {
	s := make(stringSet, len(ss))
	for _, v := range ss {
		s[v] = struct{}{}
	}
	return s
}
func hasAny(have []string, required stringSet) bool {
	for _, h := range have {
		if _, ok := required[h]; ok {
			return true
		}
	}
	return false
}
func hasAll(have []string, required stringSet) bool {
	got := toSet(have)
	for r := range required {
		if _, ok := got[r]; !ok {
			return false
		}
	}
	return true
}

func RequireScopes(scopes ...string) func(http.Handler) http.Handler {
	req := toSet(scopes)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := GetPrincipal(r)
			if !ok || p.ID == "" {
				apierrors.Unauthorized(w, "Authentication required")
				return
			}
			if !hasAny(p.Scopes, req) {
				apierrors.Forbidden(w, "Insufficient permissions: missing required scope")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAllScopes(scopes ...string) func(http.Handler) http.Handler {
	req := toSet(scopes)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := GetPrincipal(r)
			if !ok || p.ID == "" {
				apierrors.Unauthorized(w, "Authentication required")
				return
			}
			if !hasAll(p.Scopes, req) {
				apierrors.Forbidden(w, "Insufficient permissions: missing required scopes")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	req := toSet(roles)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := GetPrincipal(r)
			if !ok || p.ID == "" {
				apierrors.Unauthorized(w, "Authentication required")
				return
			}
			if !hasAny(p.Roles, req) {
				apierrors.Forbidden(w, "Insufficient permissions: missing required role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAudience(requiredAud string) func(http.Handler) http.Handler {
	req := stringSet{requiredAud: {}}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := GetPrincipal(r)
			if !ok || p.ID == "" {
				apierrors.Unauthorized(w, "Authentication required")
				return
			}
			if !hasAny(p.Audience, req) {
				apierrors.Forbidden(w, "Insufficient permissions: wrong audience")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireSubjectType(t domain.PrincipalType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := GetPrincipal(r)
			if !ok || p.ID == "" {
				apierrors.Unauthorized(w, "Authentication required")
				return
			}
			if p.Type != t {
				apierrors.Forbidden(w, "Insufficient permissions: wrong subject type")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
