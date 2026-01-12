package idrx

import (
	"net/http"
	"time"

	"github.com/igun997/idrx-go/blockchain"
)

// Client is the main entry point for the IDRX SDK using the facade pattern.
// It provides access to all API services and maintains shared configuration.
type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       AuthProvider

	// Services organized by domain/resource
	Account     *AccountService
	Transaction *TransactionService
	Blockchain  *BlockchainService // Optional blockchain service
}

// ClientOption represents a configuration option for the Client.
type ClientOption func(*Client)

// NewClient creates a new IDRX API client with the specified options.
// It follows the builder/options pattern for flexible configuration.
func NewClient(opts ...ClientOption) *Client {
	// Default configuration
	client := &Client{
		baseURL: "https://idrx.co",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Initialize services with reference to client
	client.Account = &AccountService{client: client}
	client.Transaction = &TransactionService{client: client}

	return client
}

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithAuth sets the authentication provider.
func WithAuth(auth AuthProvider) ClientOption {
	return func(c *Client) {
		c.auth = auth
	}
}

// WithBusinessAuth creates a client with business-level authentication.
// This auth type is used for general endpoints and organization management.
func WithBusinessAuth(apiKey, secretKey string) ClientOption {
	return func(c *Client) {
		c.auth = NewBusinessAuth(apiKey, secretKey)
	}
}

// WithUserAuth creates a client with user-level authentication.
// This auth type is used for user-specific operations and is obtained
// from the onboarding response.
func WithUserAuth(apiKey, secretKey string) ClientOption {
	return func(c *Client) {
		c.auth = NewUserAuth(apiKey, secretKey)
	}
}

// WithTimeout sets a custom timeout for HTTP requests.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithBlockchain enables blockchain operations with the provided private key.
// The private key should be hex-encoded (with or without 0x prefix).
func WithBlockchain(privateKeyHex string) ClientOption {
	return func(c *Client) {
		// Remove 0x prefix if present
		if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
			privateKeyHex = privateKeyHex[2:]
		}

		// Initialize blockchain client
		blockchainClient, err := blockchain.NewClient(&blockchain.ClientConfig{
			PrivateKeyHex: privateKeyHex,
			Timeout:       c.httpClient.Timeout,
		})
		if err != nil {
			// For now, we'll silently skip blockchain initialization if it fails
			// In a production environment, you might want to log this error
			return
		}

		c.Blockchain = NewBlockchainService(blockchainClient)
	}
}
