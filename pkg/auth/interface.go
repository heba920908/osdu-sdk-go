package auth

import (
	"context"
	"time"
)

// AuthProvider defines the interface that all authentication providers must implement
type AuthProvider interface {
	// GetAccessToken retrieves an access token for the given context
	GetAccessToken(ctx context.Context) (*Token, error)

	// IsTokenValid checks if the current token is still valid
	IsTokenValid() bool

	// RefreshToken attempts to refresh the current token
	RefreshToken(ctx context.Context) (*Token, error)
}

// Token represents a generic authentication token
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"-"`
	Scope        string    `json:"scope,omitempty"`
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// ProviderType represents the different authentication provider types
type ProviderType string

const (
	ProviderTypeOpenID ProviderType = "openid"
	ProviderTypeAzure  ProviderType = "azure"
)
