package domain

type PrincipalType string

const (
    PrincipalUser    PrincipalType = "user"
    PrincipalService PrincipalType = "service"
)

type Principal struct {
    Type     PrincipalType
    ID       string
    Email    string
    Roles    []string
    Scopes   []string
    ClientID string
    Audience []string
}
