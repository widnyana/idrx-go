package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	idrx "github.com/igun997/idrx-go"
	"github.com/igun997/idrx-go/blockchain"
)

func main() {
	// Read wallet private key from environment variable
	privateKey := os.Getenv("WALLET_PRIVATE_KEY")

	if privateKey == "" {
		log.Fatal("WALLET_PRIVATE_KEY must be set")
	}

	// Initialize client with blockchain capabilities
	// The private key enables direct blockchain operations
	client := idrx.NewClient(
		idrx.WithBlockchain(privateKey),
		idrx.WithTimeout(30*time.Second),
	)

	// Check if blockchain service was initialized successfully
	if client.Blockchain == nil {
		log.Fatal("Failed to initialize blockchain service. Check your private key.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wallet addresses to check
	walletsToCheck := []string{
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", // Example wallet 1
		"0x1234567890123456789012345678901234567890", // Example wallet 2
	}

	// Networks to check across
	networks := []string{
		blockchain.BaseMainnet,
		blockchain.BSCMainnet,
		blockchain.EtherlinkMainnet,
		blockchain.GnosisMainnet,
		blockchain.KaiaMainnet,
		blockchain.LiskMainnet,
		blockchain.PolygonMainnet,
		blockchain.WorldChainMainnet,
	}

	fmt.Println("=== IDRX Blacklist Verification ===")

	// Check each wallet across multiple networks
	for _, walletAddr := range walletsToCheck {
		fmt.Printf("Checking wallet: %s\n", walletAddr)
		fmt.Println("─────────────────────────────────────────────────────")

		for _, networkName := range networks {
			// Get network configuration
			networkConfig, exists := blockchain.GetNetworkConfig(networkName)
			if !exists {
				log.Printf("Warning: Network %s not found\n", networkName)
				continue
			}

			// Check blacklist status
			isBlacklisted, err := client.Blockchain.IsAddressBlacklisted(
				ctx,
				networkConfig.ChainID,
				walletAddr,
			)
			if err != nil {
				log.Printf("✗ %s (Chain %d): Error - %v\n",
					networkConfig.Name,
					networkConfig.ChainID,
					err,
				)
				continue
			}

			// Display result
			status := "✓ Not Blacklisted"
			if isBlacklisted {
				status = "✗ BLACKLISTED"
			}

			fmt.Printf("  %s (Chain %d): %s\n",
				networkConfig.Name,
				networkConfig.ChainID,
				status,
			)
		}

		fmt.Println()
	}

	// Display network information
	fmt.Println("=== Supported Networks ===")
	fmt.Println()

	for _, networkName := range networks {
		networkConfig, _ := blockchain.GetNetworkConfig(networkName)
		fmt.Printf("%s\n", networkConfig.Name)
		fmt.Printf("  Chain ID:         %d\n", networkConfig.ChainID)
		fmt.Printf("  Contract Address: %s\n", networkConfig.ContractAddress.Hex())
		fmt.Printf("  Block Time:       %s\n", networkConfig.BlockTime)
		fmt.Printf("  Gas Limit:        %d\n", networkConfig.GasLimit)
		fmt.Printf("  Max Gas Price:    %d wei\n", networkConfig.MaxGasPrice)
		fmt.Println()
	}

	fmt.Println("✓ Blacklist verification complete!")
	fmt.Println("\nℹ Blacklisted addresses cannot:")
	fmt.Println("  - Send IDRX tokens")
	fmt.Println("  - Receive IDRX tokens")
	fmt.Println("  - Participate in bridge operations")
}
