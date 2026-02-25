package blockchain

import (
	"testing"
)

func TestNetworkConstants(t *testing.T) {
	// Test that all network constants exist
	expectedNetworks := []string{
		BaseMainnet,
		PolygonMainnet,
		BSCMainnet,
		LiskMainnet,
		KaiaMainnet,
		WorldChainMainnet,
		EtherlinkMainnet,
		GnosisMainnet,
	}

	for _, networkName := range expectedNetworks {
		config, exists := GetNetworkConfig(networkName)
		if !exists {
			t.Errorf("Network %s should exist in SupportedNetworks", networkName)
			continue
		}

		if config.ChainID == 0 {
			t.Errorf("Network %s should have a valid chain ID", networkName)
		}

		if config.Name == "" {
			t.Errorf("Network %s should have a name", networkName)
		}
	}
}

func TestGetNetworkConfigByChainID(t *testing.T) {
	// Test reverse lookup by chain ID
	testCases := map[uint64]string{
		8453:  BaseMainnet,
		137:   PolygonMainnet,
		56:    BSCMainnet,
		1135:  LiskMainnet,
		8217:  KaiaMainnet,
		480:   WorldChainMainnet,
		42793: EtherlinkMainnet,
		100:   GnosisMainnet,
	}

	for chainID, expectedNetwork := range testCases {
		config, networkName, exists := GetNetworkConfigByChainID(chainID)
		if !exists {
			t.Errorf("Chain ID %d should exist", chainID)
			continue
		}

		if networkName != expectedNetwork {
			t.Errorf("Chain ID %d should map to %s, got %s", chainID, expectedNetwork, networkName)
		}

		if config.ChainID != chainID {
			t.Errorf("Config chain ID should match input: expected %d, got %d", chainID, config.ChainID)
		}
	}
}

func TestIsNetworkSupported(t *testing.T) {
	// Test valid networks
	validNetworks := []string{
		BaseMainnet,
		PolygonMainnet,
		BSCMainnet,
		LiskMainnet,
		KaiaMainnet,
		WorldChainMainnet,
		EtherlinkMainnet,
		GnosisMainnet,
	}
	for _, network := range validNetworks {
		if !IsNetworkSupported(network) {
			t.Errorf("Network %s should be supported", network)
		}
	}

	// Test invalid network
	if IsNetworkSupported("InvalidNetwork") {
		t.Error("Invalid network should not be supported")
	}
}

func TestIsChainSupported(t *testing.T) {
	// Test valid chain IDs
	validChainIDs := []uint64{8453, 137, 56, 1135, 8217, 480, 42793, 100}
	for _, chainID := range validChainIDs {
		if !IsChainSupported(chainID) {
			t.Errorf("Chain ID %d should be supported", chainID)
		}
	}

	// Test invalid chain ID
	if IsChainSupported(999999) {
		t.Error("Invalid chain ID should not be supported")
	}
}

func TestGetSupportedNetworks(t *testing.T) {
	networks := GetSupportedNetworks()
	if len(networks) == 0 {
		t.Error("Should have at least one supported network")
	}

	// Check that all expected networks are present
	expectedNetworks := map[string]bool{
		BaseMainnet:       false,
		PolygonMainnet:    false,
		BSCMainnet:        false,
		LiskMainnet:       false,
		KaiaMainnet:       false,
		WorldChainMainnet: false,
		EtherlinkMainnet:  false,
		GnosisMainnet:     false,
	}

	for _, network := range networks {
		if _, exists := expectedNetworks[network]; exists {
			expectedNetworks[network] = true
		}
	}

	for network, found := range expectedNetworks {
		if !found {
			t.Errorf("Expected network %s not found in supported networks", network)
		}
	}
}

func TestNetworkNameByChainID(t *testing.T) {
	// Test valid mappings
	testCases := map[uint64]string{
		8453:  BaseMainnet,
		137:   PolygonMainnet,
		56:    BSCMainnet,
		1135:  LiskMainnet,
		8217:  KaiaMainnet,
		480:   WorldChainMainnet,
		42793: EtherlinkMainnet,
		100:   GnosisMainnet,
	}

	for chainID, expectedName := range testCases {
		name, exists := GetNetworkNameByChainID(chainID)
		if !exists {
			t.Errorf("Chain ID %d should have a network name", chainID)
			continue
		}

		if name != expectedName {
			t.Errorf("Chain ID %d should map to %s, got %s", chainID, expectedName, name)
		}
	}

	// Test invalid chain ID
	_, exists := GetNetworkNameByChainID(999999)
	if exists {
		t.Error("Invalid chain ID should not have a network name")
	}
}

func TestGetDecimals(t *testing.T) {
	testCases := map[uint64]uint8{
		8453:  2, // Base
		137:   0, // Polygon
		56:    0, // BSC
		1135:  2, // Lisk
		8217:  2, // Kaia
		480:   2, // World Chain
		42793: 2, // Etherlink
		100:   2, // Gnosis
	}

	for chainID, expectedDecimals := range testCases {
		decimals := GetDecimals(chainID)
		if decimals != expectedDecimals {
			t.Errorf("Chain ID %d: expected decimals %d, got %d", chainID, expectedDecimals, decimals)
		}
	}
}

func TestGetDecimalsFallback(t *testing.T) {
	// Unknown chain should return fallback value of 2
	decimals := GetDecimals(999999)
	if decimals != 2 {
		t.Errorf("Unknown chain should return fallback 2, got %d", decimals)
	}
}

func TestNetworkConfigHasDecimals(t *testing.T) {
	for networkName, config := range SupportedNetworks {
		// All networks should have decimals set to 0 or 2
		if config.Decimals != 0 && config.Decimals != 2 {
			t.Errorf("Network %s has unexpected decimals value: %d", networkName, config.Decimals)
		}
	}
}
