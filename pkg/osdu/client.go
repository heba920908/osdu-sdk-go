package osdu

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
)

type osdu_api string

const (
	OsduApi osdu_api = "osdu_api"
)

// OsduApiRequest represents the OSDU API client with pluggable authentication
type OsduApiRequest struct {
	authProvider auth.AuthProvider
	osduSettings config.OsduSettings
	authSettings config.AuthSettings
}

// NewClient creates a new OSDU API client with the appropriate authentication provider
func NewClient() OsduApiRequest {
	osduSettings, _ := config.GetOsduSettings()
	authSettings, _ := config.GetAuthSettings()

	// Get the full configuration to determine provider
	configFile := config.GetConfigFile()
	cfg, _ := config.GetConfig(configFile)

	// Create authentication provider using factory
	factory := auth.NewProviderFactory()
	authProvider, err := factory.GetProviderFromConfig(cfg.OsduClient)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create auth provider: %s", err))
		// Fall back to OpenID provider
		authProvider = auth.NewOpenIDProvider(authSettings)
	}

	return OsduApiRequest{
		authProvider: authProvider,
		osduSettings: osduSettings,
		authSettings: authSettings,
	}
}

// NewClientWithProvider creates a new OSDU API client with a specific authentication provider
func NewClientWithProvider(provider auth.AuthProvider) OsduApiRequest {
	osduSettings, _ := config.GetOsduSettings()
	authSettings, _ := config.GetAuthSettings()

	return OsduApiRequest{
		authProvider: provider,
		osduSettings: osduSettings,
		authSettings: authSettings,
	}
}

// NewClientWithConfig creates a new OSDU API client with custom settings for testing
func NewClientWithConfig(provider auth.AuthProvider, osduSettings config.OsduSettings, authSettings config.AuthSettings) OsduApiRequest {
	return OsduApiRequest{
		authProvider: provider,
		osduSettings: osduSettings,
		authSettings: authSettings,
	}
}

func (a OsduApiRequest) Context() context.Context {
	return context.Background()
}

func (a OsduApiRequest) NewRequest(operation string, url string, partitionid string, body []byte) ([]byte, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, _ := http.NewRequest(operation, url, bytes.NewBuffer(body))
	headers, _ := a._build_headers_with_partition()
	req.Header = headers
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return resBody, err
	}
	res.Body.Close()
	slog.Debug("Response:")
	return resBody, nil
}

func (a OsduApiRequest) _build_headers_with_partition() (http.Header, error) {
	token, err := a.authProvider.GetAccessToken(a.Context())
	if err != nil {
		return http.Header{}, err
	}

	slog.Debug(fmt.Sprintf("Authorization Header : Bearer %s", token.AccessToken))
	return http.Header{
		"Content-Type":      {"application/json"},
		"Authorization":     {fmt.Sprintf("Bearer %s", token.AccessToken)},
		"data-partition-id": {a.osduSettings.PartitionId},
	}, nil
}

func (a OsduApiRequest) _build_headers_without_partition() (http.Header, error) {
	if a.authSettings.InternalService {
		ctx := context.Background()
		slog.InfoContext(ctx, "Internal service setup, skipping token generation")
		return http.Header{
			"Content-Type": {"application/json"},
		}, nil
	}

	token, err := a.authProvider.GetAccessToken(a.Context())
	if err != nil {
		log.Println(err)
		return http.Header{}, err
	}

	return http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token.AccessToken)},
	}, nil
}
