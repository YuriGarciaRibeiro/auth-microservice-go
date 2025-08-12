package middleware

import (
	"net/http"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
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
			println("Checking scopes:", p.Scopes, "against required:", scopes)
			if !ok || p.ID == "" {
				println("No principal found in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !hasAny(p.Scopes, req) {
				println("Principal scopes mismatch")
				http.Error(w, "forbidden (missing scope)", http.StatusForbidden)
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
			println("Checking all scopes:", p.Scopes, "against required:", scopes)
			if !ok || p.ID == "" {
				println("No principal found in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !hasAll(p.Scopes, req) {
				println("Principal scopes mismatch")
				http.Error(w, "forbidden (missing scope)", http.StatusForbidden)
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
			println("Checking roles:", p.Roles, "against required:", roles)
			if !ok || p.ID == "" {
				println("No principal found in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !hasAny(p.Roles, req) {
				println("Principal roles mismatch")
				http.Error(w, "forbidden (missing role)", http.StatusForbidden)
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
			println("Checking audience:", p.Audience, "against required:", requiredAud)
			if !ok || p.ID == "" {
				println("No principal found in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !hasAny(p.Audience, req) {
				println("Principal audience mismatch")
				http.Error(w, "forbidden (wrong audience)", http.StatusForbidden)
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
			println("Checking subject type:", p.Type, "against required:", t)
			if !ok || p.ID == "" {
				println("No principal found in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if p.Type != t {
				println("Principal type mismatch")
				http.Error(w, "forbidden (subject type)", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
