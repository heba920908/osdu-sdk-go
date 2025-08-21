# Azure Authentication Implementation Guide

This guide explains how to use the Azure authentication provider in the OSDU SDK Go.

## Overview

The Azure authentication provider supports multiple authentication methods:

1. **Azure Managed Identity (Pod Identity)** - For Azure AKS workloads
2. **Service Principal with Client Secret** - For application authentication
3. **OAuth2 Fallback** - Traditional OAuth2 flow when Azure SDK is not available

## Configuration

### Service Principal Authentication

```yaml
osdu:
  provider: azure
  auth:
    clientId: "your-app-id"
    clientSecret: "your-client-secret"
    tenantId: "your-tenant-id"
    scopes: "https://graph.microsoft.com/.default"
    grantType: client_credentials
    podAuthEnabled: false
```

### Azure Managed Identity (Pod Identity)

```yaml
osdu:
  provider: azure
  auth:
    podAuthEnabled: true
    tenantId: "your-tenant-id"
    scopes: "https://management.azure.com/.default"
```

### Environment Variables

Set these environment variables for secure credential management:

```bash
export OSDU_AUTH_CLIENT_ID="your-app-id"
export OSDU_AUTH_CLIENT_SECRET="your-client-secret"
export OSDU_AUTH_TENANT_ID="your-tenant-id"
```

## Usage Examples

### Factory Pattern (Recommended)

```go
import (
    "github.com/heba920908/osdu-sdk-go/pkg/osdu"
)

// Automatically uses Azure provider if configured
client := osdu.NewClient()
```

### Direct Provider Usage

```go
import (
    "context"
    "github.com/heba920908/osdu-sdk-go/pkg/auth"
    "github.com/heba920908/osdu-sdk-go/pkg/config"
    "github.com/heba920908/osdu-sdk-go/pkg/osdu"
)

authSettings, _ := config.GetAuthSettings()
azureProvider, err := auth.NewAzureProvider(authSettings)
if err != nil {
    log.Fatal(err)
}

// Create client with Azure provider
client := osdu.NewClientWithProvider(azureProvider)

// Get token directly
token, err := azureProvider.GetAccessToken(context.Background())
if err != nil {
    log.Fatal(err)
}
```

## Authentication Flow

### 1. Pod Identity Flow (podAuthEnabled: true)

1. Uses `azidentity.NewDefaultAzureCredential()`
2. Automatically handles Managed Identity, Azure CLI, etc.
3. Gets token via Azure SDK with specified scopes and tenant

### 2. Service Principal Flow

1. Uses `azidentity.NewClientSecretCredential()` 
2. Authenticates with client ID, secret, and tenant ID
3. Gets token via Azure SDK

### 3. OAuth2 Fallback Flow

1. Falls back to traditional OAuth2 when Azure SDK credentials are not available
2. Uses HTTP POST to token endpoint
3. Supports client credentials and refresh token grant types

## Scopes

The provider supports various Azure scopes:

- `https://graph.microsoft.com/.default` - Microsoft Graph (default)
- `https://management.azure.com/.default` - Azure Management
- `https://vault.azure.net/.default` - Azure Key Vault
- Custom application scopes

Multiple scopes can be specified separated by spaces:
```yaml
scopes: "https://graph.microsoft.com/.default https://management.azure.com/.default"
```

## Error Handling

The provider handles various error scenarios:

- **Missing credentials**: Returns descriptive error messages
- **Token expiration**: Automatically refreshes tokens
- **Network failures**: Provides retry capabilities through Azure SDK
- **Invalid configuration**: Validates required fields

## Token Caching

- Tokens are cached in memory per provider instance
- Automatic token refresh when expired
- Thread-safe token operations

## Azure AKS Deployment

For Azure AKS with Pod Identity:

1. Enable Pod Identity on your AKS cluster
2. Create Azure Identity and binding
3. Set `podAuthEnabled: true` in configuration
4. Deploy your application

```yaml
# Example AKS deployment snippet
spec:
  template:
    metadata:
      labels:
        aadpodidbinding: "your-identity-binding"
    spec:
      containers:
      - name: osdu-app
        env:
        - name: CONFIG_FILE
          value: "/config/azure-config.yaml"
```

## Troubleshooting

### Common Issues

1. **"failed to create Azure credential"**
   - Check client ID, secret, and tenant ID are correct
   - Verify the service principal has required permissions

2. **"Token request failed"**
   - Verify scopes are correct for your use case
   - Check network connectivity to Azure endpoints

3. **"Pod identity not found"**
   - Ensure Pod Identity is properly configured in AKS
   - Verify identity binding is correct

### Debug Logging

Enable debug logging to troubleshoot issues:

```go
import "log/slog"

// Set debug level logging
slog.SetLogLoggerLevel(slog.LevelDebug)
```

## Performance Considerations

- Token caching reduces authentication overhead
- Azure SDK handles connection pooling automatically
- Managed Identity has lower latency than service principal authentication

## Security Best Practices

1. **Use Managed Identity when possible** - Most secure option
2. **Store secrets in Azure Key Vault** - Don't hardcode credentials
3. **Use least privilege scopes** - Only request necessary permissions
4. **Rotate credentials regularly** - Implement credential rotation
5. **Monitor authentication logs** - Track usage and failures
