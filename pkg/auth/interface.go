package auth

import (
	"context"
	"encoding/json"
	"strings"
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
	Scopes       []string  `json:"-"`
}

// UnmarshalJSON custom unmarshaler to handle scope field as string or array
func (t *Token) UnmarshalJSON(data []byte) error {
	// Define a temporary struct with scope as string
	type Alias Token
	aux := &struct {
		ScopeString string `json:"scope,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse space-separated scope string into slice
	if aux.ScopeString != "" {
		t.Scopes = strings.Fields(aux.ScopeString)
	}

	return nil
}

// MarshalJSON custom marshaler to output scope as space-separated string
func (t Token) MarshalJSON() ([]byte, error) {
	type Alias Token
	return json.Marshal(&struct {
		ScopeString string `json:"scope,omitempty"`
		Alias
	}{
		ScopeString: strings.Join(t.Scopes, " "),
		Alias:       (Alias)(t),
	})
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
