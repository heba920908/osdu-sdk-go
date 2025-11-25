package v2

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
)

// OsduClient is the main interface for interacting with OSDU services
type OsduClient interface {
	// Register returns the register service interface
	Register() RegisterService

	// Context returns the context for requests
	Context() context.Context

	// Logger returns the logger instance
	Logger() *slog.Logger
}

// HTTPClient interface for making HTTP requests (useful for mocking)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client implements the OsduClient interface
type Client struct {
	authProvider auth.AuthProvider
	osduSettings config.OsduSettings
	httpClient   HTTPClient
	ctx          context.Context
	logger       *slog.Logger
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithAuthProvider sets a custom auth provider
func WithAuthProvider(provider auth.AuthProvider) ClientOption {
	return func(c *Client) {
		c.authProvider = provider
	}
}

// WithOsduSettings sets custom OSDU settings
func WithOsduSettings(settings config.OsduSettings) ClientOption {
	return func(c *Client) {
		c.osduSettings = settings
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient HTTPClient) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithContext sets a custom context
func WithContext(ctx context.Context) ClientOption {
	return func(c *Client) {
		c.ctx = ctx
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient creates a new OSDU client with the default configuration
// If WithOsduSettings and WithHTTPClient options are provided, it skips loading from config files
func NewClient(opts ...ClientOption) (OsduClient, error) {
	// Create a temporary client to check if required options are provided
	tempClient := &Client{}
	for _, opt := range opts {
		opt(tempClient)
	}

	// If settings and HTTP client are provided, skip config loading
	skipConfig := tempClient.osduSettings.RegisterUrl != "" && tempClient.httpClient != nil

	var osduSettings config.OsduSettings
	var authProvider auth.AuthProvider

	if !skipConfig {
		var err error
		osduSettings, err = config.GetOsduSettings()
		if err != nil {
			return nil, fmt.Errorf("failed to get OSDU settings: %w", err)
		}

		authSettings, err := config.GetAuthSettings()
		if err != nil {
			return nil, fmt.Errorf("failed to get auth settings: %w", err)
		}

		// Get the full configuration to determine provider
		configFile := config.GetConfigFile()
		cfg, err := config.GetConfig(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to get config: %w", err)
		}

		// Create authentication provider using factory
		factory := auth.NewProviderFactory()
		authProvider, err = factory.GetProviderFromConfig(cfg.OsduClient)
		if err != nil {
			// Use default logger for initial setup
			logger := tempClient.logger
			if logger == nil {
				logger = slog.Default()
			}
			logger.Warn("Failed to create auth provider from config, falling back to OpenID",
				"error", err)
			authProvider = auth.NewOpenIDProvider(authSettings)
		}
	} else {
		// Use provided settings
		osduSettings = tempClient.osduSettings
		authProvider = tempClient.authProvider
	}

	// Configure default HTTP client with TLS settings if not provided
	defaultHTTPClient := tempClient.httpClient
	if defaultHTTPClient == nil {
		defaultHTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	// Configure default logger if not provided
	defaultLogger := tempClient.logger
	if defaultLogger == nil {
		defaultLogger = slog.Default()
	}

	client := &Client{
		authProvider: authProvider,
		osduSettings: osduSettings,
		httpClient:   defaultHTTPClient,
		ctx:          context.Background(),
		logger:       defaultLogger,
	}

	// Apply options again to override any defaults
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Context returns the context for the client
func (c *Client) Context() context.Context {
	return c.ctx
}

// Logger returns the logger instance
func (c *Client) Logger() *slog.Logger {
	return c.logger
}

// Register returns the register service
func (c *Client) Register() RegisterService {
	return &registerService{client: c}
}

// buildHeaders builds HTTP headers with authorization
func (c *Client) buildHeaders(includeAuth bool) (http.Header, error) {
	headers := http.Header{
		"Content-Type": {"application/json"},
	}

	if includeAuth {
		token, err := c.authProvider.GetAccessToken(c.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}

		if token != nil && token.AccessToken != "" {
			headers.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		}
	}

	return headers, nil
}

// buildHeadersWithPartition builds HTTP headers with authorization and partition ID
func (c *Client) buildHeadersWithPartition() (http.Header, error) {
	if c.osduSettings.PartitionId == "" {
		return nil, fmt.Errorf("partition ID is required")
	}

	headers, err := c.buildHeaders(true)
	if err != nil {
		return nil, err
	}

	headers.Set("data-partition-id", c.osduSettings.PartitionId)
	c.logger.Debug("Partition header set",
		"data-partition-id", c.osduSettings.PartitionId)

	return headers, nil
}

// doRequest performs an HTTP request with the given parameters
func (c *Client) doRequest(method, url string, body []byte, includeAuth, includePartition bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var headers http.Header
	if includePartition {
		headers, err = c.buildHeadersWithPartition()
	} else {
		headers, err = c.buildHeaders(includeAuth)
	}
	if err != nil {
		return nil, err
	}

	req.Header = headers

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// readResponse reads the response body and handles errors
func readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
