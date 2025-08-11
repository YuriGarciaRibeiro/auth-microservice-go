package usecase

import (
	"errors"
	"strings"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type ClientCredentialsUseCase struct {
	Repo domain.ClientRepository
}

func NewClientCredentialsUseCase(repo domain.ClientRepository) *ClientCredentialsUseCase {
	return &ClientCredentialsUseCase{Repo: repo}
}

type ClientCredentialsInput struct {
	ClientID string
	Secret   string
	Scopes   []string
	Audience []string
}

func (uc *ClientCredentialsUseCase) Execute(in ClientCredentialsInput) (domain.Principal, error) {
	c, err := uc.Repo.FindByClientID(in.ClientID)
	if err != nil || c == nil || !c.Active {
		return domain.Principal{}, errors.New("invalid client")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(c.SecretHash), []byte(in.Secret)); err != nil {
		return domain.Principal{}, errors.New("invalid client")
	}

	allowedScopes := trimAll(c.AllowedScopes)
	allowedAud := trimAll(c.AllowedAudience)

	if len(in.Audience) == 0 || !containsOne(allowedAud, trimAll(in.Audience)) {
		return domain.Principal{}, errors.New("invalid audience")
	}

	effScopes := unique(intersect(trimAll(in.Scopes), allowedScopes))

	return domain.Principal{
		Type:     domain.PrincipalService,
		ID:       c.ID,
		ClientID: c.ClientID,
		Scopes:   effScopes,
		Audience: trimAll(in.Audience),
	}, nil
}

func intersect(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, x := range b {
		set[x] = struct{}{}
	}
	var out []string
	for _, y := range a {
		if _, ok := set[y]; ok {
			out = append(out, y)
		}
	}
	return out
}

func containsAll(allowed, required []string) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[a] = struct{}{}
	}
	for _, r := range required {
		if _, ok := set[r]; !ok {
			return false
		}
	}
	return true
}

func containsOne(allowed, requested []string) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[a] = struct{}{}
	}
	for _, r := range requested {
		if _, ok := set[r]; ok {
			return true
		}
	}
	return false
}

func unique(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func trimAll(in []string) []string {
	var out []string
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}
