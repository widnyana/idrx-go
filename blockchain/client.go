package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/igun997/idrx-go/contracts"
)

// Client represents a multi-chain blockchain client for IDRX operations
type Client struct {
	// clients maps chain ID to ethereum clients with failover support
	clients map[uint64][]*ethclient.Client

	// contracts maps chain ID to IDRX contract instances
	contracts map[uint64]*contracts.IDRX

	// privateKey for signing transactions
	privateKey *ecdsa.PrivateKey

	// publicKey derived from private key
	publicKey *ecdsa.PublicKey

	// address derived from public key
	address common.Address

	// mutex for thread-safe operations
	mutex sync.RWMutex

	// timeout for RPC requests
	timeout time.Duration
}

// ClientConfig represents configuration for the blockchain client
type ClientConfig struct {
	PrivateKeyHex string        // Hex-encoded private key
	Timeout       time.Duration // RPC request timeout
}

// NewClient creates a new blockchain client with multi-chain support
func NewClient(config *ClientConfig) (*Client, error) {
	// Parse private key
	privateKey, err := crypto.HexToECDSA(config.PrivateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Derive public key and address
	publicKeyInterface := privateKey.Public()
	publicKey, ok := publicKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA public key")
	}
	address := crypto.PubkeyToAddress(*publicKey)

	client := &Client{
		clients:    make(map[uint64][]*ethclient.Client),
		contracts:  make(map[uint64]*contracts.IDRX),
		privateKey: privateKey,
		publicKey:  publicKey,
		address:    address,
		timeout:    config.Timeout,
	}

	if client.timeout == 0 {
		client.timeout = 30 * time.Second
	}

	// Initialize connections to all supported networks
	if err := client.initializeNetworks(); err != nil {
		return nil, fmt.Errorf("failed to initialize networks: %w", err)
	}

	return client, nil
}

// initializeNetworks connects to all supported blockchain networks
func (c *Client) initializeNetworks() error {
	for _, networkConfig := range SupportedNetworks {
		// Skip if contract not deployed on this network
		if networkConfig.ContractAddress == (common.Address{}) {
			continue
		}

		var ethClients []*ethclient.Client
		var contractInstance *contracts.IDRX

		// Connect to each RPC endpoint for failover support
		for _, rpcEndpoint := range networkConfig.RPCEndpoints {
			ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
			ethClient, err := ethclient.DialContext(ctx, rpcEndpoint)
			cancel()

			if err != nil {
				// Log error but continue with other endpoints
				fmt.Printf("Failed to connect to %s (%s): %v\n", networkConfig.Name, rpcEndpoint, err)
				continue
			}

			ethClients = append(ethClients, ethClient)

			// Create contract instance using the first successful connection
			if contractInstance == nil {
				contractInstance, err = contracts.NewIDRX(networkConfig.ContractAddress, ethClient)
				if err != nil {
					return fmt.Errorf("failed to create contract instance for %s: %w", networkConfig.Name, err)
				}
			}
		}

		if len(ethClients) == 0 {
			return fmt.Errorf("failed to connect to any RPC endpoint for %s", networkConfig.Name)
		}

		c.clients[networkConfig.ChainID] = ethClients
		c.contracts[networkConfig.ChainID] = contractInstance
	}

	return nil
}

// GetClient returns an ethereum client for the specified chain ID with failover
func (c *Client) GetClient(chainID uint64) (*ethclient.Client, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	clients, exists := c.clients[chainID]
	if !exists {
		return nil, fmt.Errorf("chain ID %d not supported or not connected", chainID)
	}

	// Return the first available client (simple failover strategy)
	for _, client := range clients {
		// Test if client is still connected
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := client.ChainID(ctx)
		cancel()

		if err == nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("no available clients for chain ID %d", chainID)
}

// GetContract returns the IDRX contract instance for the specified chain ID
func (c *Client) GetContract(chainID uint64) (*contracts.IDRX, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	contract, exists := c.contracts[chainID]
	if !exists {
		return nil, fmt.Errorf("IDRX contract not available for chain ID %d", chainID)
	}

	return contract, nil
}

// GetClientByNetwork returns an ethereum client for the specified network name with failover
func (c *Client) GetClientByNetwork(networkName string) (*ethclient.Client, error) {
	networkConfig, exists := GetNetworkConfig(networkName)
	if !exists {
		return nil, fmt.Errorf("network %s not supported", networkName)
	}
	return c.GetClient(networkConfig.ChainID)
}

// GetContractByNetwork returns the IDRX contract instance for the specified network name
func (c *Client) GetContractByNetwork(networkName string) (*contracts.IDRX, error) {
	networkConfig, exists := GetNetworkConfig(networkName)
	if !exists {
		return nil, fmt.Errorf("network %s not supported", networkName)
	}
	return c.GetContract(networkConfig.ChainID)
}

// GetAddress returns the wallet address associated with this client
func (c *Client) GetAddress() common.Address {
	return c.address
}

// CreateTransactor creates a transactor for signing transactions on the specified chain
func (c *Client) CreateTransactor(ctx context.Context, chainID uint64) (*bind.TransactOpts, error) {
	client, err := c.GetClient(chainID)
	if err != nil {
		return nil, err
	}

	// Get chain ID from client to ensure it matches
	clientChainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	if clientChainID.Uint64() != chainID {
		return nil, fmt.Errorf("client chain ID %d does not match requested chain ID %d", clientChainID.Uint64(), chainID)
	}

	// Create transactor with the private key
	transactor, err := bind.NewKeyedTransactorWithChainID(c.privateKey, clientChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Get network config for gas settings
	networkConfig, _, exists := GetNetworkConfigByChainID(chainID)
	if exists {
		transactor.GasLimit = networkConfig.GasLimit
		if networkConfig.MaxGasPrice > 0 {
			transactor.GasPrice = new(big.Int).SetUint64(networkConfig.MaxGasPrice)
		}
	}

	return transactor, nil
}

// EstimateGas estimates gas for a contract method call
func (c *Client) EstimateGas(ctx context.Context, chainID uint64, msg ethereum.CallMsg) (uint64, error) {
	client, err := c.GetClient(chainID)
	if err != nil {
		return 0, err
	}

	return client.EstimateGas(ctx, msg)
}

// WaitForTransaction waits for a transaction to be mined and returns the receipt
func (c *Client) WaitForTransaction(ctx context.Context, chainID uint64, txHash common.Hash) (*types.Receipt, error) {
	client, err := c.GetClient(chainID)
	if err != nil {
		return nil, err
	}

	// Get network config for appropriate timeout
	networkConfig, _, exists := GetNetworkConfigByChainID(chainID)
	timeout := 5 * time.Minute // Default timeout
	if exists {
		// Calculate timeout based on block time (allow for ~10 blocks)
		timeout = networkConfig.BlockTime * 10
		if timeout < time.Minute {
			timeout = time.Minute
		}
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(networkConfig.BlockTime / 2) // Check twice per block
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction %s", txHash.Hex())
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(timeoutCtx, txHash)
			if err == nil {
				return receipt, nil
			}
			// Continue waiting if transaction not found yet
		}
	}
}

// Close closes all ethereum client connections
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, clients := range c.clients {
		for _, client := range clients {
			client.Close()
		}
	}
}
