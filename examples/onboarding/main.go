package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	idrx "github.com/igun997/idrx-go"
	"github.com/igun997/idrx-go/models"
)

func main() {
	// Read credentials from environment variables
	apiKey := os.Getenv("IDRX_API_KEY")
	apiSecret := os.Getenv("IDRX_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("IDRX_API_KEY and IDRX_API_SECRET must be set")
	}

	// Initialize client with business authentication
	// Business auth is used for organization-level operations like user onboarding
	client := idrx.NewClient(
		idrx.WithBusinessAuth(apiKey, apiSecret),
		idrx.WithTimeout(30*time.Second),
	)

	// Create context with timeout for the request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Open ID document file
	// In production, this would be the user's ID card/passport scan
	idFile, err := os.Open("sample-id.jpg")
	if err != nil {
		log.Printf("Warning: Could not open sample ID file: %v", err)
		log.Printf("In production, provide actual ID document file")
		// For demo purposes, we'll continue with nil file
		// In real usage, this would fail at API level
	}
	if idFile != nil {
		defer idFile.Close()
	}

	// Prepare onboarding request
	onboardingReq := &models.OnboardingRequest{
		Email:    "user@example.com",
		Fullname: "John Doe",
		Address:  "123 Main Street, Jakarta, Indonesia",
		IDNumber: "ID123456789",
		IDFile:   idFile,
	}

	fmt.Println("Starting user onboarding...")
	fmt.Printf("Email: %s\n", onboardingReq.Email)
	fmt.Printf("Name: %s\n", onboardingReq.Fullname)

	// Submit onboarding request
	resp, err := client.Account.Onboard(ctx, onboardingReq)
	if err != nil {
		// Handle API errors
		if apiErr, ok := err.(*models.APIError); ok {
			log.Fatalf("API Error [%d]: %s", apiErr.StatusCode, apiErr.Message)
		}
		log.Fatalf("Onboarding failed: %v", err)
	}

	// Success! Display user credentials
	fmt.Println("\n✓ Onboarding successful!")
	fmt.Println("\n=== User Credentials ===")
	fmt.Printf("API Key:    %s\n", resp.APIKey)
	fmt.Printf("API Secret: %s\n", resp.APISecret)
	fmt.Printf("User ID:    %d\n", resp.ID)
	fmt.Printf("Full Name:  %s\n", resp.Fullname)
	fmt.Printf("Created:    %s\n", resp.CreatedAt.Format(time.RFC3339))

	fmt.Println("\nℹ Store these credentials securely!")
	fmt.Println("These credentials are used for user-level operations:")
	fmt.Println("- Minting IDRX")
	fmt.Println("- Redemption requests")
	fmt.Println("- Bank account management")
	fmt.Println("- Transaction history")
}
