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

	// Example 3: Using Azure provider
	fmt.Println("\n=== Azure Provider Example ===")
	azureProvider, err := auth.NewAzureProvider(authSettings)
	if err != nil {
		log.Printf("Failed to create Azure provider: %s", err)
	} else {
		clientWithAzure := osdu.NewClientWithProvider(azureProvider)
		fmt.Printf("Client created with Azure provider: %+v\n", clientWithAzure)

		// Test Azure token retrieval (this will work with proper config)
		fmt.Println("\n=== Testing Azure Token Retrieval ===")
		azureToken, err := azureProvider.GetAccessToken(context.Background())
		if err != nil {
			log.Printf("Failed to get Azure token: %s", err)
		} else {
			fmt.Printf("Azure token retrieved successfully. Length: %d\n", len(azureToken.AccessToken))
			fmt.Printf("Azure token type: %s\n", azureToken.TokenType)
			fmt.Printf("Azure token expires at: %s\n", azureToken.ExpiresAt)
		}
	}

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
