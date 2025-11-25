package v2_test

import (
	"context"
	"net/http"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
)

// mockHTTPClient is a mock implementation of HTTPClient for testing
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// mockAuthProvider is a mock implementation of AuthProvider for testing
type mockAuthProvider struct {
	GetAccessTokenFunc func(ctx context.Context) (*auth.Token, error)
	IsTokenValidFunc   func() bool
	RefreshTokenFunc   func(ctx context.Context) (*auth.Token, error)
}

func (m *mockAuthProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
	if m.GetAccessTokenFunc != nil {
		return m.GetAccessTokenFunc(ctx)
	}
	return &auth.Token{AccessToken: "mock-token"}, nil
}

func (m *mockAuthProvider) IsTokenValid() bool {
	if m.IsTokenValidFunc != nil {
		return m.IsTokenValidFunc()
	}
	return true
}

func (m *mockAuthProvider) RefreshToken(ctx context.Context) (*auth.Token, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx)
	}
	return &auth.Token{AccessToken: "mock-refreshed-token"}, nil
}
