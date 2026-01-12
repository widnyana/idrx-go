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
	// Read user credentials from environment variables
	// These are obtained from the onboarding response
	apiKey := os.Getenv("IDRX_API_KEY")
	apiSecret := os.Getenv("IDRX_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("IDRX_API_KEY and IDRX_API_SECRET must be set")
	}

	// Initialize client with user authentication
	// User auth is used for transaction operations
	client := idrx.NewClient(
		idrx.WithUserAuth(apiKey, apiSecret),
		idrx.WithTimeout(30*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 1: Query current swap rates
	fmt.Println("=== Step 1: Query Swap Rates ===")
	ratesReq := models.NewIDRXRatesRequest("100000", string(models.ChainPolygon))

	rates, err := client.Transaction.GetRates(ctx, ratesReq)
	if err != nil {
		log.Fatalf("Failed to get rates: %v", err)
	}

	fmt.Printf("Buy Amount: %s\n", rates.BuyAmount)
	fmt.Printf("From:       %s\n", rates.FromToken)
	fmt.Printf("To:         %s\n", rates.ToToken)
	fmt.Printf("Rate:       %s\n", rates.Price)
	fmt.Printf("Chain ID:   %s\n\n", rates.ChainID)

	// Step 2: Create mint request
	fmt.Println("=== Step 2: Create Mint Request ===")
	mintReq := &models.MintRequest{
		ToBeMinted:               "100000",                                     // 100K IDRX
		DestinationWalletAddress: "0x61Ba72c10983Def54AA7F93a334837BAA5d93396", // change this to your destination wallet address
		NetworkChainID:           string(models.ChainPolygon),                  // Polygon
		ExpiryPeriod:             3600,                                         // 1 hour
	}

	mintResp, err := client.Transaction.MintRequest(ctx, mintReq)
	if err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			log.Fatalf("API Error [%d]: %s", apiErr.StatusCode, apiErr.Message)
		}
		log.Fatalf("Mint request failed: %v", err)
	}

	fmt.Printf("✓ Mint request created\n")
	fmt.Printf("Payment URL:  %s\n", mintResp.PaymentURL)
	fmt.Printf("Amount:       %s IDRX\n", mintResp.AdjustedAmount)
	fmt.Printf("Expires at:   %s\n", mintResp.ExpiresAt.Format(time.RFC3339))
	if mintResp.TransactionID != "" {
		fmt.Printf("Transaction:  %s\n", mintResp.TransactionID)
	}
	fmt.Println()

	// Step 3: Add bank account for redemption
	fmt.Println("=== Step 3: Add Bank Account ===")
	bankReq := &models.AddBankAccountRequest{
		BankAccountNumber: "1234567890",
		BankCode:          "BCA",
	}

	bankResp, err := client.Account.AddBankAccount(ctx, bankReq)
	if err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			log.Fatalf("API Error [%d]: %s", apiErr.StatusCode, apiErr.Message)
		}
		log.Fatalf("Add bank account failed: %v", err)
	}

	fmt.Printf("✓ Bank account added\n")
	fmt.Printf("Account ID:     %d\n", bankResp.ID)
	fmt.Printf("Bank:           %s (%s)\n", bankResp.BankName, bankResp.BankCode)
	fmt.Printf("Account Number: %s\n", bankResp.BankAccountNumber)
	fmt.Printf("Deposit Wallet: %s\n", bankResp.DepositWalletAddress.Address)
	fmt.Printf("Chain:          %s (%s)\n\n", bankResp.DepositWalletAddress.ChainName, bankResp.DepositWalletAddress.ChainID)

	// Step 4: List bank accounts
	fmt.Println("=== Step 4: List Bank Accounts ===")
	accounts, err := client.Account.GetBankAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get bank accounts: %v", err)
	}

	fmt.Printf("Found %d bank account(s)\n", len(accounts))
	for i, acc := range accounts {
		fmt.Printf("%d. %s - %s (%s)\n", i+1, acc.BankName, acc.BankAccountNumber, acc.DepositWalletAddress.Address)
	}
	fmt.Println()

	// Step 5: Create redemption request
	fmt.Println("=== Step 5: Create Redemption Request ===")
	redeemReq := &models.RedeemRequest{
		TxHash:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		NetworkChainID:  string(models.ChainPolygon),
		AmountTransfer:  "50000", // 50K IDRX
		BankAccount:     "1234567890",
		BankCode:        "BCA",
		BankName:        "Bank Central Asia",
		BankAccountName: "John Doe",
		WalletAddress:   "0x61Ba72c10983Def54AA7F93a334837BAA5d93396",
		Notes:           "Redemption request",
	}

	redeemResp, err := client.Transaction.RedeemRequest(ctx, redeemReq)
	if err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			log.Fatalf("API Error [%d]: %s", apiErr.StatusCode, apiErr.Message)
		}
		log.Fatalf("Redeem request failed: %v", err)
	}

	fmt.Printf("✓ Redemption request created\n")
	fmt.Printf("Transaction ID:   %s\n", redeemResp.TransactionID)
	fmt.Printf("Status:           %s\n", redeemResp.Status)
	if redeemResp.ReferenceNumber != "" {
		fmt.Printf("Reference:        %s\n", redeemResp.ReferenceNumber)
	}
	if redeemResp.ProcessingTime != "" {
		fmt.Printf("Processing Time:  %s\n", redeemResp.ProcessingTime)
	}
	if !redeemResp.EstimatedComplete.IsZero() {
		fmt.Printf("Est. Complete:    %s\n", redeemResp.EstimatedComplete.Format(time.RFC3339))
	}
	fmt.Println()

	fmt.Println("✓ Complete workflow finished successfully!")
	fmt.Println("\nℹ Next steps:")
	fmt.Println("1. Complete payment at the mint payment URL")
	fmt.Println("2. Monitor redemption status via transaction history")
	fmt.Println("3. Bank transfer will be processed according to fee tier")
}
