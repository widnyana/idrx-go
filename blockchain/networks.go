// Package blockchain provides network configuration and constants for IDRX token operations across multiple blockchains.
package blockchain

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// NetworkConfig represents configuration for a blockchain network
type NetworkConfig struct {
	ChainID         uint64
	Name            string
	RPCEndpoints    []string // Multiple RPC endpoints for failover
	ContractAddress common.Address
	BlockTime       time.Duration
	GasLimit        uint64
	MaxGasPrice     uint64 // in wei
	IsTestnet       bool
	Decimals        uint8 // Token decimals for this deployment
}

// Network identifier constants for better type safety
const (
	BaseMainnet       = "BaseMainnet"
	PolygonMainnet    = "PolygonMainnet"
	BSCMainnet        = "BSCMainnet"
	LiskMainnet       = "LiskMainnet"
	KaiaMainnet       = "KaiaMainnet"
	WorldChainMainnet = "WorldChainMainnet"
	EtherlinkMainnet  = "EtherlinkMainnet"
	GnosisMainnet     = "GnosisMainnet"
)

// Chain ID constants for supported networks
const (
	BaseChainID       = 8453
	PolygonChainID    = 137
	BSCChainID        = 56
	LiskChainID       = 1135
	KaiaChainID       = 8217
	WorldChainChainID = 480
	EtherlinkChainID  = 42793
	GnosisChainID     = 100
)

// SupportedNetworks contains all IDRX supported blockchain networks
var SupportedNetworks = map[string]*NetworkConfig{
	// Base Mainnet (Primary network)
	BaseMainnet: {
		ChainID:         BaseChainID,
		Name:            "Base",
		RPCEndpoints:    []string{"https://mainnet.base.org", "https://base.llamarpc.com"},
		ContractAddress: common.HexToAddress("0x18Bc5bcC660cf2B9cE3cd51a404aFe1a0cBD3C22"),
		BlockTime:       2 * time.Second,
		GasLimit:        3000000,
		MaxGasPrice:     10000000000, // 10 Gwei
		IsTestnet:       false,
		Decimals:        2,
	},
	// Polygon Mainnet
	PolygonMainnet: {
		ChainID:         PolygonChainID,
		Name:            "Polygon",
		RPCEndpoints:    []string{"https://polygon-rpc.com", "https://polygon.llamarpc.com"},
		ContractAddress: common.HexToAddress("0x649a2DA7B28E0D54c13D5eFf95d3A660652742cC"),
		BlockTime:       2 * time.Second,
		GasLimit:        3000000,
		MaxGasPrice:     50000000000, // 50 Gwei
		IsTestnet:       false,
		Decimals:        0,
	},
	// BNB Smart Chain
	BSCMainnet: {
		ChainID:         BSCChainID,
		Name:            "BNB Smart Chain",
		RPCEndpoints:    []string{"https://bsc-dataseed.binance.org", "https://binance.llamarpc.com"},
		ContractAddress: common.HexToAddress("0x649a2DA7B28E0D54c13D5eFf95d3A660652742cC"),
		BlockTime:       3 * time.Second,
		GasLimit:        3000000,
		MaxGasPrice:     5000000000, // 5 Gwei (BSC gas price: 0.05 Gwei as of 2025)
		IsTestnet:       false,
		Decimals:        0,
	},
	// Lisk Mainnet
	LiskMainnet: {
		ChainID:         LiskChainID,
		Name:            "Lisk",
		RPCEndpoints:    []string{"https://rpc.api.lisk.com"},
		ContractAddress: common.HexToAddress("0x18Bc5bcC660cf2B9cE3cd51a404aFe1a0cBD3C22"),
		BlockTime:       2 * time.Second,
		GasLimit:        3000000,
		MaxGasPrice:     1000000000, // 1 Gwei
		IsTestnet:       false,
		Decimals:        2,
	},
	KaiaMainnet: {
		ChainID:         KaiaChainID,
		Name:            "Kaia",
		RPCEndpoints:    []string{"https://public-en.node.kaia.io"},
		ContractAddress: common.HexToAddress("0x18Bc5bcC660cf2B9cE3cd51a404aFe1a0cBD3C22"),
		BlockTime:       2 * time.Second,
		GasLimit:        3000000,
		MaxGasPrice:     1000000000, // 1 Gwei
		IsTestnet:       false,
		Decimals:        2,
	},
	WorldChainMainnet: {
		ChainID:         WorldChainChainID,
		Name:            "World Chain",
		RPCEndpoints:    []string{"https://worldchain-mainnet.g.alchemy.com/public"},
		ContractAddress: common.HexToAddress("0x18Bc5bcC660cf2B9cE3cd51a404aFe1a0cBD3C22"),
		BlockTime:       2 * time.Second,
		GasLimit:        3000000,
		MaxGasPrice:     1000000000, // 1 Gwei
		IsTestnet:       false,
		Decimals:        2,
	},
	EtherlinkMainnet: {
		ChainID:         EtherlinkChainID,
		Name:            "Etherlink",
		RPCEndpoints:    []string{"https://node.mainnet.etherlink.com"},
		ContractAddress: common.HexToAddress("0x18bc5bcc660cf2b9ce3cd51a404afe1a0cbd3c22"),
		BlockTime:       2 * time.Second,
		GasLimit:        30000000,   // 30M (Dionysus upgrade allows up to 30M gas units)
		MaxGasPrice:     1000000000, // 1 Gwei
		IsTestnet:       false,
		Decimals:        2,
	},
	GnosisMainnet: {
		ChainID:         GnosisChainID,
		Name:            "Gnosis",
		RPCEndpoints:    []string{"https://rpc.gnosischain.com", "https://0xrpc.io/gno"},
		ContractAddress: common.HexToAddress("0x18bc5bcc660cf2b9ce3cd51a404afe1a0cbd3c22"),
		BlockTime:       5 * time.Second,
		GasLimit:        3000000,
		MaxGasPrice:     500000000, // 0.5 Gwei (current gas price ~0.2 Gwei)
		IsTestnet:       false,
		Decimals:        2,
	},
}

// GetNetworkConfig returns the network configuration for a given network name
func GetNetworkConfig(networkName string) (*NetworkConfig, bool) {
	config, exists := SupportedNetworks[networkName]
	return config, exists
}

// GetNetworkConfigByChainID returns the network configuration for a given chain ID
func GetNetworkConfigByChainID(chainID uint64) (*NetworkConfig, string, bool) {
	for networkName, config := range SupportedNetworks {
		if config.ChainID == chainID {
			return config, networkName, true
		}
	}
	return nil, "", false
}

// GetSupportedNetworks returns a list of all supported network names
func GetSupportedNetworks() []string {
	networks := make([]string, 0, len(SupportedNetworks))
	for networkName := range SupportedNetworks {
		networks = append(networks, networkName)
	}
	return networks
}

// GetSupportedChainIDs returns a list of all supported chain IDs
func GetSupportedChainIDs() []uint64 {
	chainIDs := make([]uint64, 0, len(SupportedNetworks))
	for _, config := range SupportedNetworks {
		chainIDs = append(chainIDs, config.ChainID)
	}
	return chainIDs
}

// IsNetworkSupported checks if a network name is supported
func IsNetworkSupported(networkName string) bool {
	_, exists := SupportedNetworks[networkName]
	return exists
}

// IsChainSupported checks if a chain ID is supported
func IsChainSupported(chainID uint64) bool {
	for _, config := range SupportedNetworks {
		if config.ChainID == chainID {
			return true
		}
	}
	return false
}

// GetMainnetNetworks returns only mainnet networks (non-testnet)
func GetMainnetNetworks() map[string]*NetworkConfig {
	mainnets := make(map[string]*NetworkConfig)
	for networkName, config := range SupportedNetworks {
		if !config.IsTestnet {
			mainnets[networkName] = config
		}
	}
	return mainnets
}

// GetNetworkNameByChainID returns the network name for a given chain ID
func GetNetworkNameByChainID(chainID uint64) (string, bool) {
	for networkName, config := range SupportedNetworks {
		if config.ChainID == chainID {
			return networkName, true
		}
	}
	return "", false
}

// GetDecimals returns the token decimals for a given chain ID.
// Returns 2 as default if chain not found (majority of chains use 2 decimals).
func GetDecimals(chainID uint64) uint8 {
	config, _, exists := GetNetworkConfigByChainID(chainID)
	if !exists {
		return 2 // fallback to majority decimal places
	}
	return config.Decimals
}
