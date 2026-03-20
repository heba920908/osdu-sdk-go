package main

import (
	"context"
	"fmt"
	"log"

	azureprovider "github.com/heba920908/osdu-sdk-go/examples/auth_azure_mod_example"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
	"github.com/heba920908/osdu-sdk-go/pkg/osdu"
)

func main() {
	fmt.Println("=== Azure Authentication Provider Example ===")

	// Load configuration
	authSettings, err := config.GetAuthSettings()
	if err != nil {
		log.Printf("Warning: Could not load auth settings: %v", err)
		log.Println("Using example configuration instead...")

		// Example configuration for demonstration
		authSettings = config.AuthSettings{
			SdkAuth:  true, // Use Azure Managed Identity
			TenantId: "your-tenant-id",
			Scopes:   []string{"https://your-osdu-instance.com/.default"},
		}
	}

	// Create Azure provider
	fmt.Println("Creating Azure authentication provider...")
	azureProvider, err := azureprovider.NewAzureProvider(authSettings)
	if err != nil {
		log.Fatalf("Failed to create Azure provider: %v", err)
	}
	fmt.Println("✓ Azure provider created successfully")

	// Create OSDU client with Azure provider
	fmt.Println("\nCreating OSDU client with Azure authentication...")
	client := osdu.NewClientWithProvider(azureProvider)
	fmt.Printf("✓ Client created: %+v\n", client)

	// Test token retrieval
	fmt.Println("\n=== Testing Token Retrieval ===")
	ctx := context.Background()
	token, err := azureProvider.GetAccessToken(ctx)
	if err != nil {
		log.Printf("⚠ Failed to get token (expected if not running in Azure): %v", err)
		fmt.Println("\nThis is normal if you're not running in an Azure environment")
		fmt.Println("with proper credentials configured.")
	} else {
		fmt.Printf("✓ Token retrieved successfully!\n")
		fmt.Printf("  Token Type: %s\n", token.TokenType)
		fmt.Printf("  Token Length: %d characters\n", len(token.AccessToken))
		fmt.Printf("  Expires At: %s\n", token.ExpiresAt)
		fmt.Printf("  Scopes: %v\n", token.Scopes)

		// Test token validity
		if azureProvider.IsTokenValid() {
			fmt.Println("  ✓ Token is valid")
		} else {
			fmt.Println("  ⚠ Token is not valid")
		}
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nTo use this in your project:")
	fmt.Println("1. Import the module:")
	fmt.Println("   go get github.com/heba920908/osdu-sdk-go/examples/auth_azure_mod_example")
	fmt.Println("2. Use in your code:")
	fmt.Println("   import azureprovider \"github.com/heba920908/osdu-sdk-go/examples/auth_azure_mod_example\"")
	fmt.Println("   provider, _ := azureprovider.NewAzureProvider(authSettings)")
	fmt.Println("\nOr copy azure_provider.go directly into your project for full customization.")
}
