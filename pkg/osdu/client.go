package osdu

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
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
	}
}

// NewClientWithProvider creates a new OSDU API client with a specific authentication provider
func NewClientWithProvider(provider auth.AuthProvider) OsduApiRequest {
	osduSettings, _ := config.GetOsduSettings()

	return OsduApiRequest{
		authProvider: provider,
		osduSettings: osduSettings,
	}
}

// NewClientWithConfig creates a new OSDU API client with custom settings for testing
func NewClientWithConfig(provider auth.AuthProvider, osduSettings config.OsduSettings) OsduApiRequest {
	return OsduApiRequest{
		authProvider: provider,
		osduSettings: osduSettings,
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
	slog.Debug(fmt.Sprintf("Partition Header - data-partition-id : %s", a.osduSettings.PartitionId))
	if len(a.osduSettings.PartitionId) < 2 {
		return http.Header{}, errors.New("invalid partition id")
	}
	token, err := a.authProvider.GetAccessToken(a.Context())
	if err != nil {
		return http.Header{}, err
	}

	slog.Debug(fmt.Sprintf("Authorization Header - Authorization: Bearer %s", token.AccessToken))
	return http.Header{
		"Content-Type":      {"application/json"},
		"Authorization":     {fmt.Sprintf("Bearer %s", token.AccessToken)},
		"data-partition-id": {a.osduSettings.PartitionId},
	}, nil
}

func (a OsduApiRequest) _build_headers_without_partition() (http.Header, error) {
	token, err := a.authProvider.GetAccessToken(a.Context())
	if err != nil {
		log.Println(err)
		return http.Header{}, err
	}

	// If no access token is provided (empty), return headers without authorization
	if token == nil || token.AccessToken == "" {
		ctx := context.Background()
		slog.InfoContext(ctx, "No access token provided, proceeding without authorization")
		return http.Header{
			"Content-Type": {"application/json"},
		}, nil
	}

	return http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token.AccessToken)},
	}, nil
}

// HttpRequestWithoutPartition makes an HTTP request without the data-partition-id header
func (a OsduApiRequest) HttpRequestWithoutPartition(method, url string, body []byte) (*http.Response, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	headers, err := a._build_headers_without_partition()
	if err != nil {
		return nil, err
	}
	req.Header = headers

	client := http.Client{}
	return client.Do(req)
}
