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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
)

// AzureProvider implements the AuthProvider interface for Azure Active Directory
type AzureProvider struct {
	config       config.AuthSettings
	credential   azcore.TokenCredential
	currentToken *Token
	scopes       []string
}

// NewAzureProvider creates a new Azure authentication provider
func NewAzureProvider(authConfig config.AuthSettings) (*AzureProvider, error) {
	// Parse scopes from the configuration
	scopes := []string{authConfig.Scopes}
	if authConfig.Scopes == "" {
		// Default to Microsoft Graph scope if none specified
		scopes = []string{"https://graph.microsoft.com/.default"}
	} else if strings.Contains(authConfig.Scopes, " ") {
		// Split multiple scopes
		scopes = strings.Fields(authConfig.Scopes)
	}

	var credential azcore.TokenCredential
	var err error

	// Create Azure credential if Pod Auth is enabled or we have client credentials
	if authConfig.PodAuthEnabled {
		// Use default Azure credential (managed identity, Azure CLI, etc.)
		credential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create default Azure credential: %w", err)
		}
	} else if authConfig.ClientId != "" && authConfig.ClientSecret != "" && authConfig.TenantId != "" {
		// Use client secret credential for service principal authentication
		credential, err = azidentity.NewClientSecretCredential(
			authConfig.TenantId,
			authConfig.ClientId,
			authConfig.ClientSecret,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create client secret credential: %w", err)
		}
	}

	return &AzureProvider{
		config:     authConfig,
		credential: credential,
		scopes:     scopes,
	}, nil
}

// GetAccessToken retrieves an access token using Azure SDK or OAuth2
func (p *AzureProvider) GetAccessToken(ctx context.Context) (*Token, error) {
	slog.DebugContext(ctx, fmt.Sprintf("Azure Auth: Expire - %v", p.currentToken))

	if p.IsTokenValid() {
		slog.DebugContext(ctx, fmt.Sprintf("Azure token still active ... %s > %s", p.currentToken.ExpiresAt, time.Now()))
		return p.currentToken, nil
	}

	slog.InfoContext(ctx, "Azure - Generating new token")

	// Check if Pod Auth is enabled (Azure Managed Identity)
	if p.config.PodAuthEnabled {
		return p.getTokenWithAzureSDK(ctx)
	}

	// Check if we have Azure credentials for SDK
	if p.credential != nil {
		return p.getTokenWithAzureSDK(ctx)
	}

	// Fall back to OAuth2 flow if no Azure credentials
	return p.getTokenWithOAuth2(ctx)
}

// getTokenWithAzureSDK uses Azure SDK for authentication
func (p *AzureProvider) getTokenWithAzureSDK(ctx context.Context) (*Token, error) {
	slog.InfoContext(ctx, "Azure - Using Azure SDK authentication")

	tokenRequestOptions := policy.TokenRequestOptions{
		Scopes: p.scopes,
	}

	// Add tenant ID if available
	if p.config.TenantId != "" {
		tokenRequestOptions.TenantID = p.config.TenantId
	}

	accessToken, err := p.credential.GetToken(ctx, tokenRequestOptions)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error while obtaining Azure token: %s", err))
		return nil, fmt.Errorf("failed to get Azure access token: %w", err)
	}

	// Convert Azure token to our generic Token structure
	token := &Token{
		AccessToken: accessToken.Token,
		TokenType:   "Bearer",
		ExpiresAt:   accessToken.ExpiresOn,
		ExpiresIn:   int(time.Until(accessToken.ExpiresOn).Seconds()),
		Scope:       strings.Join(p.scopes, " "),
	}

	// Store the token
	p.currentToken = token

	slog.InfoContext(ctx, fmt.Sprintf("Azure SDK - Done | AT: %d", len(token.AccessToken)))
	return token, nil
}

// getTokenWithOAuth2 uses OAuth2 flow for authentication (fallback)
func (p *AzureProvider) getTokenWithOAuth2(ctx context.Context) (*Token, error) {
	slog.InfoContext(ctx, "Azure - Using OAuth2 flow authentication")

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

	formVals.Set("scope", strings.Join(p.scopes, " "))
	if len(p.config.ClientSecret) > 0 {
		formVals.Set("client_secret", p.config.ClientSecret)
	}

	slog.InfoContext(ctx, fmt.Sprintf("Trying: %s", p.config.TokenUrl))
	slog.InfoContext(ctx, fmt.Sprintf("grant_type: %s", p.config.GrantType))
	slog.InfoContext(ctx, fmt.Sprintf("client_id: %s", p.config.ClientId))
	slog.InfoContext(ctx, fmt.Sprintf("scope: %s", p.scopes))

	response, err := http.PostForm(p.config.TokenUrl, formVals)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error while obtaining token: %s", err))
		return nil, fmt.Errorf("error while trying to get tokens: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode > 302 {
		slog.ErrorContext(ctx, fmt.Sprintf("Unexpected auth code: %d", response.StatusCode))
		return nil, fmt.Errorf("unexpected auth code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error while trying to read token json body: %w", err)
	}

	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, fmt.Errorf("error while trying to parse token json body: %w", err)
	}

	// Set expiration time (fallback to 60 minutes if not provided)
	if token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	} else {
		token.ExpiresAt = time.Now().Add(60 * time.Minute)
	}

	// Store the token
	p.currentToken = &token

	slog.InfoContext(ctx, fmt.Sprintf("Azure OAuth2 - Done - code: %s | AT: %d", response.Status, len(token.AccessToken)))
	return &token, nil
}

// IsTokenValid checks if the current token is valid and not expired
func (p *AzureProvider) IsTokenValid() bool {
	if p.currentToken == nil {
		return false
	}
	return len(p.currentToken.AccessToken) > 5 && !p.currentToken.IsExpired()
}

// RefreshToken attempts to refresh the current token
func (p *AzureProvider) RefreshToken(ctx context.Context) (*Token, error) {
	// For Azure SDK, just request a new token (SDK handles refresh automatically)
	if p.config.PodAuthEnabled || p.credential != nil {
		return p.getTokenWithAzureSDK(ctx)
	}

	// For OAuth2, try refresh token flow if available
	if p.currentToken != nil && p.currentToken.RefreshToken != "" {
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

		return p.getTokenWithOAuth2(ctx)
	}

	// Fall back to getting a new token
	return p.GetAccessToken(ctx)
}

// GetScopes returns the configured scopes for testing and debugging purposes
func (p *AzureProvider) GetScopes() []string {
	// Return a copy to prevent external modification
	scopes := make([]string, len(p.scopes))
	copy(scopes, p.scopes)
	return scopes
}
