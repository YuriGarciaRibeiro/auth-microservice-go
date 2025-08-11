package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
)

type ctxKey string

const principalCtxKey ctxKey = "auth.principal"

func GetPrincipal(r *http.Request) (domain.Principal, bool) {
	v := r.Context().Value(principalCtxKey)
	if v == nil {
		return domain.Principal{}, false
	}
	if p, ok := v.(domain.Principal); ok {
		return p, true
	}
	return domain.Principal{}, false
}

func MustPrincipal(w http.ResponseWriter, r *http.Request) domain.Principal {
	p, ok := GetPrincipal(r)
	if !ok || p.ID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return domain.Principal{} // caller should return after this
	}
	return p
}

func Authn(tokens domain.TokenService) func(http.Handler) http.Handler {
	// We receive the TokenService here to avoid global state and ease testing.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "missing or malformed Authorization header", http.StatusUnauthorized)
				return
			}
			access := strings.TrimPrefix(auth, "Bearer ")
			if access == "" {
				http.Error(w, "missing access token", http.StatusUnauthorized)
				return
			}

			claims, err := tokens.VerifyAccess(access)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}
			principal := domain.Principal{
				Type:     claims.SubjectType, 
				ID:       claims.SubjectID,  
				Email:    claims.Email,
				Roles:    claims.Roles,
				Scopes:   claims.Scopes,
				ClientID: claims.ClientID,
				Audience: claims.Audience,
			}
			
			ctx := context.WithValue(r.Context(), principalCtxKey, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
