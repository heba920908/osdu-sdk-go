package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/heba920908/osdu-sdk-go/pkg/config"
)

// OpenIDProvider implements the AuthProvider interface for OpenID Connect/OAuth2
type OpenIDProvider struct {
	config       config.AuthSettings
	currentToken *Token
}

// NewOpenIDProvider creates a new OpenID authentication provider
func NewOpenIDProvider(authConfig config.AuthSettings) *OpenIDProvider {
	return &OpenIDProvider{
		config: authConfig,
	}
}

// GetAccessToken retrieves an access token using OpenID Connect/OAuth2
func (p *OpenIDProvider) GetAccessToken(ctx context.Context) (*Token, error) {
	if p.IsTokenValid() {
		slog.DebugContext(ctx, fmt.Sprintf("Token still active ... %s > %s", p.currentToken.ExpiresAt, time.Now()))
		return p.currentToken, nil
	}

	slog.InfoContext(ctx, "OpenID - Generating new token")

	formVals := url.Values{}
	formVals.Set("client_id", p.config.ClientId)
	formVals.Set("grant_type", p.config.GrantType)

	if p.config.GrantType == "refresh_token" {
		refreshToken := p.config.RefreshToken
		if p.currentToken != nil && p.currentToken.RefreshToken != "" {
			refreshToken = p.currentToken.RefreshToken
		}
		formVals.Set("refresh_token", refreshToken)
	}

	formVals.Set("scope", strings.Join(p.config.Scopes, " "))
	if len(p.config.ClientSecret) > 0 {
		formVals.Set("client_secret", p.config.ClientSecret)
	}

	slog.InfoContext(ctx, fmt.Sprintf("Trying: %s", p.config.TokenUrl))
	slog.InfoContext(ctx, fmt.Sprintf("grant_type: %s", p.config.GrantType))
	slog.InfoContext(ctx, fmt.Sprintf("client_id: %s", p.config.ClientId))
	slog.InfoContext(ctx, fmt.Sprintf("scope: %s", strings.Join(p.config.Scopes, " ")))

	response, err := http.PostForm(p.config.TokenUrl, formVals)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error while obtaining token: %s", err))
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode > 302 {
		slog.ErrorContext(ctx, fmt.Sprintf("Unexpected auth code: %d", response.StatusCode))
		return nil, fmt.Errorf("unexpected auth code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, err
	}

	// Set expiration time
	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	// Store the token
	p.currentToken = &token

	slog.InfoContext(ctx, fmt.Sprintf("OpenID - Done - code: %s | AT: %d", response.Status, len(token.AccessToken)))
	return &token, nil
}

// IsTokenValid checks if the current token is valid and not expired
func (p *OpenIDProvider) IsTokenValid() bool {
	if p.currentToken == nil {
		return false
	}
	return len(p.currentToken.AccessToken) > 5 && !p.currentToken.IsExpired()
}

// RefreshToken attempts to refresh the current token
func (p *OpenIDProvider) RefreshToken(ctx context.Context) (*Token, error) {
	if p.currentToken == nil || p.currentToken.RefreshToken == "" {
		return p.GetAccessToken(ctx)
	}

	// Update config to use refresh token grant type temporarily
	originalGrantType := p.config.GrantType
	originalRefreshToken := p.config.RefreshToken

	p.config.GrantType = "refresh_token"
	p.config.RefreshToken = p.currentToken.RefreshToken

	// Restore original values after refresh
	defer func() {
		p.config.GrantType = originalGrantType
		p.config.RefreshToken = originalRefreshToken
	}()

	return p.GetAccessToken(ctx)
}
