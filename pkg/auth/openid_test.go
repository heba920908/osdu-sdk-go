package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that OpenIDProvider implements the AuthProvider interface
func TestOpenIDProviderInterface(t *testing.T) {
	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     "https://example.com/token",
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	var provider auth.AuthProvider = auth.NewOpenIDProvider(authConfig)
	assert.NotNil(t, provider)
}

func TestNewOpenIDProvider(t *testing.T) {
	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     "https://example.com/token",
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)

	assert.NotNil(t, provider)
	assert.False(t, provider.IsTokenValid()) // No token initially
}

func TestOpenIDProvider_GetAccessToken_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		// Parse form data
		err := r.ParseForm()
		require.NoError(t, err)

		assert.Equal(t, "test-client-id", r.FormValue("client_id"))
		assert.Equal(t, "test-client-secret", r.FormValue("client_secret"))
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.Equal(t, "openid profile", r.FormValue("scope"))

		// Return a successful token response
		tokenResponse := map[string]interface{}{
			"access_token": "mock-access-token-12345",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "openid profile",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	token, err := provider.GetAccessToken(ctx)

	require.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "mock-access-token-12345", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.Equal(t, 3600, token.ExpiresIn)
	assert.Equal(t, "openid profile", token.Scope)
	assert.True(t, time.Now().Before(token.ExpiresAt))
	assert.False(t, token.IsExpired())
}

func TestOpenIDProvider_GetAccessToken_WithRefreshToken(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)

		assert.Equal(t, "test-client-id", r.FormValue("client_id"))
		assert.Equal(t, "test-client-secret", r.FormValue("client_secret"))
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.Equal(t, "openid profile", r.FormValue("scope"))

		// Return a successful token response with refresh token
		tokenResponse := map[string]interface{}{
			"access_token":  "mock-access-token-12345",
			"refresh_token": "mock-refresh-token-67890",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "openid profile",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	token, err := provider.GetAccessToken(ctx)

	require.NoError(t, err)
	assert.Equal(t, "mock-refresh-token-67890", token.RefreshToken)
}

func TestOpenIDProvider_GetAccessToken_CachedToken(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		tokenResponse := map[string]interface{}{
			"access_token": "mock-access-token-12345",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "openid profile",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	// First call should hit the server
	token1, err := provider.GetAccessToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call should use cached token
	token2, err := provider.GetAccessToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount) // Should not increment
	assert.Equal(t, token1.AccessToken, token2.AccessToken)
}

func TestOpenIDProvider_GetAccessToken_HttpError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_client"}`))
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "invalid-client-id",
		ClientSecret: "invalid-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	token, err := provider.GetAccessToken(ctx)

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "unexpected auth code: 401")
}

func TestOpenIDProvider_GetAccessToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"invalid": json`)) // Invalid JSON
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	token, err := provider.GetAccessToken(ctx)

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestOpenIDProvider_GetAccessToken_NetworkError(t *testing.T) {
	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     "http://invalid-url-that-should-not-exist.local",
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	token, err := provider.GetAccessToken(ctx)

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestOpenIDProvider_GetAccessToken_WithoutClientSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)

		// Verify client_secret is not sent when empty
		assert.Equal(t, "", r.FormValue("client_secret"))
		assert.Equal(t, "test-client-id", r.FormValue("client_id"))

		tokenResponse := map[string]interface{}{
			"access_token": "mock-access-token-12345",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "openid profile",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "", // Empty client secret
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	token, err := provider.GetAccessToken(ctx)

	require.NoError(t, err)
	assert.NotNil(t, token)
}

func TestOpenIDProvider_IsTokenValid(t *testing.T) {
	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     "https://example.com/token",
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)

	// No token initially
	assert.False(t, provider.IsTokenValid())

	// Set up a mock server for getting a token
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenResponse := map[string]interface{}{
			"access_token": "mock-access-token-12345",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "openid profile",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	provider = auth.NewOpenIDProvider(config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	})

	// Get a valid token
	ctx := context.Background()
	token, err := provider.GetAccessToken(ctx)
	require.NoError(t, err)
	require.NotNil(t, token)

	// Should be valid now
	assert.True(t, provider.IsTokenValid())

	// Manually expire the token
	token.ExpiresAt = time.Now().Add(-time.Hour)

	// Should be invalid now
	assert.False(t, provider.IsTokenValid())
}

func TestOpenIDProvider_RefreshToken_WithRefreshToken(t *testing.T) {
	var requestCount int
	var lastGrantType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		err := r.ParseForm()
		require.NoError(t, err)

		lastGrantType = r.FormValue("grant_type")

		var tokenResponse map[string]interface{}

		if lastGrantType == "refresh_token" {
			// Refresh token request
			assert.Equal(t, "mock-refresh-token-67890", r.FormValue("refresh_token"))
			tokenResponse = map[string]interface{}{
				"access_token":  "new-access-token-98765",
				"refresh_token": "new-refresh-token-54321",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"scope":         "openid profile",
			}
		} else {
			// Initial token request
			tokenResponse = map[string]interface{}{
				"access_token":  "mock-access-token-12345",
				"refresh_token": "mock-refresh-token-67890",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"scope":         "openid profile",
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	// Get initial token
	token, err := provider.GetAccessToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "mock-access-token-12345", token.AccessToken)
	assert.Equal(t, "mock-refresh-token-67890", token.RefreshToken)
	assert.Equal(t, 1, requestCount)

	// Manually expire the token to force refresh
	token.ExpiresAt = time.Now().Add(-time.Hour)

	// Refresh the token
	refreshedToken, err := provider.RefreshToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "new-access-token-98765", refreshedToken.AccessToken)
	assert.Equal(t, "new-refresh-token-54321", refreshedToken.RefreshToken)
	assert.Equal(t, 2, requestCount)
	assert.Equal(t, "refresh_token", lastGrantType)
}

func TestOpenIDProvider_RefreshToken_WithoutRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenResponse := map[string]interface{}{
			"access_token": "mock-access-token-12345",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "openid profile",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	// Refresh token without any existing token should fall back to GetAccessToken
	token, err := provider.RefreshToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "mock-access-token-12345", token.AccessToken)
}

func TestOpenIDProvider_RefreshToken_ConfigGrantType(t *testing.T) {
	var requestCount int
	var lastGrantType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		err := r.ParseForm()
		require.NoError(t, err)

		lastGrantType = r.FormValue("grant_type")

		tokenResponse := map[string]interface{}{
			"access_token": fmt.Sprintf("token-%d", requestCount),
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "openid profile",
		}

		if lastGrantType == "refresh_token" {
			tokenResponse["refresh_token"] = "new-refresh-token"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenUrl:     server.URL,
		GrantType:    "refresh_token",
		RefreshToken: "config-refresh-token",
		Scopes:       "openid profile",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	// First call should use refresh_token grant type from config
	token, err := provider.GetAccessToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "token-1", token.AccessToken)
	assert.Equal(t, "refresh_token", lastGrantType)

	// Manually expire the token to force refresh
	token.ExpiresAt = time.Now().Add(-time.Hour)

	// Refresh should also use refresh_token grant type
	refreshedToken, err := provider.RefreshToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "token-2", refreshedToken.AccessToken)
	assert.Equal(t, "refresh_token", lastGrantType)
}

func TestOpenIDProvider_GrantTypes(t *testing.T) {
	testCases := []struct {
		name      string
		grantType string
	}{
		{"client_credentials", "client_credentials"},
		{"authorization_code", "authorization_code"},
		{"refresh_token", "refresh_token"},
		{"password", "password"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := r.ParseForm()
				require.NoError(t, err)

				assert.Equal(t, tc.grantType, r.FormValue("grant_type"))

				tokenResponse := map[string]interface{}{
					"access_token": "mock-access-token",
					"token_type":   "Bearer",
					"expires_in":   3600,
					"scope":        "openid profile",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tokenResponse)
			}))
			defer server.Close()

			authConfig := config.AuthSettings{
				ClientId:     "test-client-id",
				ClientSecret: "test-client-secret",
				TokenUrl:     server.URL,
				GrantType:    tc.grantType,
				RefreshToken: "test-refresh-token", // For refresh_token grant type
				Scopes:       "openid profile",
			}

			provider := auth.NewOpenIDProvider(authConfig)
			ctx := context.Background()

			token, err := provider.GetAccessToken(ctx)
			require.NoError(t, err)
			assert.Equal(t, "mock-access-token", token.AccessToken)
		})
	}
}

func TestOpenIDProvider_UrlEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)

		// Verify that special characters are properly encoded
		assert.Equal(t, "client with spaces", r.FormValue("client_id"))
		assert.Equal(t, "secret&with=special?chars", r.FormValue("client_secret"))
		assert.Equal(t, "openid profile email", r.FormValue("scope"))

		tokenResponse := map[string]interface{}{
			"access_token": "mock-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer server.Close()

	authConfig := config.AuthSettings{
		ClientId:     "client with spaces",
		ClientSecret: "secret&with=special?chars",
		TokenUrl:     server.URL,
		GrantType:    "client_credentials",
		Scopes:       "openid profile email",
	}

	provider := auth.NewOpenIDProvider(authConfig)
	ctx := context.Background()

	token, err := provider.GetAccessToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "mock-access-token", token.AccessToken)
}
