// internal/domain/client.go
package domain

type Client struct {
	ID              string
	ClientID        string
	SecretHash      string
	Name            string
	AllowedScopes   []string 
	AllowedAudience []string 
	Active          bool
}

type ClientRepository interface {
	FindByClientID(clientID string) (*Client, error)
}
