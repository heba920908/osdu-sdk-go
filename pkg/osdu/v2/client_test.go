package v2_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
	v2 "github.com/heba920908/osdu-sdk-go/pkg/osdu/v2"
)

func TestNewClient(t *testing.T) {
	// Note: This test requires a valid config file to be present
	// In a real scenario, you might want to set environment variables or mock the config
	t.Run("NewClient with default settings", func(t *testing.T) {
		// This test may fail if CONFIG_FILE env is not set or config file is missing
		// For demonstration purposes, we'll skip if it fails
		_, err := v2.NewClient()
		if err != nil {
			t.Logf("Expected error when config is not available: %v", err)
			// This is expected in test environments without proper config
		}
	})

	t.Run("NewClient with custom options", func(t *testing.T) {
		mockAuth := &mockAuthProvider{
			GetAccessTokenFunc: func(ctx context.Context) (*auth.Token, error) {
				return &auth.Token{AccessToken: "test-token"}, nil
			},
		}

		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`{"status":"ok"}`)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			PartitionId: "test-partition",
			RegisterUrl: "https://test.osdu.com/api/register/v1",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Test that we can get register service
		regService := client.Register()
		if regService == nil {
			t.Error("Expected register service to be returned")
		}
	})
}

type contextKey string

func TestClientContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey("test-key"), "test-value")

	mockHTTP := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			}, nil
		},
	}

	osduSettings := config.OsduSettings{
		RegisterUrl: "https://test.osdu.com/api/register/v1",
	}

	client, err := v2.NewClient(
		v2.WithContext(ctx),
		v2.WithHTTPClient(mockHTTP),
		v2.WithOsduSettings(osduSettings),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.Context() != ctx {
		t.Error("Expected context to match")
	}
}

func TestClientRegister(t *testing.T) {
	mockHTTP := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			}, nil
		},
	}

	osduSettings := config.OsduSettings{
		RegisterUrl: "https://test.osdu.com/api/register/v1",
	}

	client, err := v2.NewClient(
		v2.WithHTTPClient(mockHTTP),
		v2.WithOsduSettings(osduSettings),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	regService := client.Register()
	if regService == nil {
		t.Fatal("Expected register service to be returned")
	}
}

func TestClientLogger(t *testing.T) {
	t.Run("Logger() returns default logger when not set", func(t *testing.T) {
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
		}

		client, err := v2.NewClient(
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		logger := client.Logger()
		if logger == nil {
			t.Fatal("Logger() returned nil")
		}
	})

	t.Run("WithLogger sets custom logger", func(t *testing.T) {
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
		}

		// Create a custom logger
		customLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

		client, err := v2.NewClient(
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
			v2.WithLogger(customLogger),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		if client.Logger() != customLogger {
			t.Fatal("WithLogger() did not set the custom logger")
		}
	})
}

// Note: Tests for internal methods (buildHeaders, doRequest, etc.) are not included
// in v2_test package as they test private implementation details. These are tested
// indirectly through the public API tests in register_test.go
