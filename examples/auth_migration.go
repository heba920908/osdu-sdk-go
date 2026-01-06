package main

import (
	"context"
	"fmt"
	"log"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
	"github.com/heba920908/osdu-sdk-go/pkg/osdu"
)

func main() {
	// Example 1: Using factory pattern (recommended)
	fmt.Println("=== Using Factory Pattern ===")
	client := osdu.NewClient()
	fmt.Printf("Client created with factory pattern: %+v\n", client)

	// Example 2: Using specific provider directly
	fmt.Println("\n=== Using OpenID Provider Directly ===")
	authSettings, _ := config.GetAuthSettings()
	openidProvider := auth.NewOpenIDProvider(authSettings)
	clientWithOpenID := osdu.NewClientWithProvider(openidProvider)
	fmt.Printf("Client created with OpenID provider: %+v\n", clientWithOpenID)

	// Example 3: Custom Authentication Provider
	fmt.Println("\n=== Custom Provider Example ===")
	fmt.Println("For Azure authentication, see: examples/auth_azure_mod_example/")
	fmt.Println("You can implement custom providers by implementing the auth.AuthProvider interface")
	fmt.Println("This keeps the main library lightweight while allowing full customization")

	// Example 4: Testing token retrieval
	fmt.Println("\n=== Testing Token Retrieval ===")
	token, err := openidProvider.GetAccessToken(context.Background())
	if err != nil {
		log.Printf("Failed to get token: %s", err)
	} else {
		fmt.Printf("Token retrieved successfully. Length: %d\n", len(token.AccessToken))
		fmt.Printf("Token type: %s\n", token.TokenType)
		fmt.Printf("Expires at: %s\n", token.ExpiresAt)
	}
}
