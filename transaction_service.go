package idrx

import (
	"context"
	"fmt"

	"github.com/igun997/idrx-go/models"
)

// TransactionService handles transaction operations including minting, redeeming,
// rate queries, and transaction history.
type TransactionService struct {
	client *Client
}

// MintRequest creates a request to mint IDRX or other stablecoins.
// This operation requires user-level authentication.
func (s *TransactionService) MintRequest(ctx context.Context, req *models.MintRequest) (*models.MintResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("mint request cannot be nil")
	}

	// Validate required fields
	if err := s.validateMintRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var response models.APIResponse[models.MintResponse]
	if err := s.client.doRequest(ctx, "POST", "/api/transaction/mint-request", req, &response); err != nil {
		return nil, fmt.Errorf("failed to create mint request: %w", err)
	}

	return &response.Data, nil
}

// RedeemRequest creates a request to redeem IDRX to a bank account.
// This operation requires user-level authentication.
func (s *TransactionService) RedeemRequest(ctx context.Context, req *models.RedeemRequest) (*models.RedeemResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("redeem request cannot be nil")
	}

	// Validate required fields
	if err := s.validateRedeemRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var response models.APIResponse[models.RedeemResponse]
	if err := s.client.doRequest(ctx, "POST", "/api/transaction/redeem-request", req, &response); err != nil {
		return nil, fmt.Errorf("failed to create redeem request: %w", err)
	}

	return &response.Data, nil
}

// BridgeRequest creates a request for token bridging operations.
// This operation requires user-level authentication.
func (s *TransactionService) BridgeRequest(ctx context.Context, req *models.BridgeRequest) (*models.BridgeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("bridge request cannot be nil")
	}

	// Validate required fields
	if err := s.validateBridgeRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var response models.APIResponse[models.BridgeResponse]
	if err := s.client.doRequest(ctx, "POST", "/api/transaction/bridge-request", req, &response); err != nil {
		return nil, fmt.Errorf("failed to create bridge request: %w", err)
	}

	return &response.Data, nil
}

// GetRates retrieves current swap rates between IDRX and other tokens.
// Either idrxAmount or usdtAmount must be provided, but not both.
func (s *TransactionService) GetRates(ctx context.Context, req *models.RatesRequest) (*models.RatesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("rates request cannot be nil")
	}

	// Validate that exactly one amount is provided
	if err := s.validateRatesRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Build query parameters
	queryParams, err := s.client.buildQueryParams(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build query parameters: %w", err)
	}

	path := "/api/transaction/rates"
	if queryParams != "" {
		path += "?" + queryParams
	}

	var response models.APIResponse[models.RatesResponse]
	if err := s.client.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get rates: %w", err)
	}

	return &response.Data, nil
}

// GetMethods retrieves the list of supported banks with codes and transfer limits.
func (s *TransactionService) GetMethods(ctx context.Context) ([]models.Bank, error) {
	var response models.BanksResponse

	if err := s.client.doRequest(ctx, "GET", "/api/transaction/method", nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get methods: %w", err)
	}

	return response.Data, nil
}

// GetBanks is a wrapper for GetMethods to fix vague IDRX terms
func (s *TransactionService) GetBanks(ctx context.Context) ([]models.Bank, error) {
	return s.GetMethods(ctx)
}

// GetAdditionalFees retrieves additional fees for different transaction types.
func (s *TransactionService) GetAdditionalFees(ctx context.Context, req *models.FeesRequest) ([]models.Fee, error) {
	if req == nil {
		return nil, fmt.Errorf("fees request cannot be nil")
	}

	// Validate required fields
	if req.FeeType == "" {
		return nil, fmt.Errorf("fee type is required")
	}

	// Build query parameters
	queryParams, err := s.client.buildQueryParams(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build query parameters: %w", err)
	}

	path := "/api/transaction/get-additional-fees"
	if queryParams != "" {
		path += "?" + queryParams
	}

	var response models.FeesResponse
	if err := s.client.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get additional fees: %w", err)
	}

	return response.Data, nil
}

// GetTransactionHistory retrieves the user's transaction history with filtering and pagination.
// This operation requires user-level authentication.
func (s *TransactionService) GetTransactionHistory(
	ctx context.Context,
	req *models.TransactionHistoryRequest,
) (*models.TransactionHistoryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("transaction history request cannot be nil")
	}

	// Validate required fields
	if err := s.validateTransactionHistoryRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Build query parameters
	queryParams, err := s.client.buildQueryParams(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build query parameters: %w", err)
	}

	path := "/api/transaction/user-transaction-history"
	if queryParams != "" {
		path += "?" + queryParams
	}

	var response models.TransactionHistoryResponse
	if err := s.client.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	return &response, nil
}

// validateMintRequest validates the mint request fields.
func (s *TransactionService) validateMintRequest(req *models.MintRequest) error {
	if req.ToBeMinted == "" {
		return fmt.Errorf("toBeMinted is required")
	}
	if req.DestinationWalletAddress == "" {
		return fmt.Errorf("destinationWalletAddress is required")
	}
	if req.NetworkChainID == "" {
		return fmt.Errorf("networkChainId is required")
	}
	if req.ExpiryPeriod <= 0 {
		return fmt.Errorf("expiryPeriod must be greater than 0")
	}
	return nil
}

// validateRedeemRequest validates the redeem request fields.
func (s *TransactionService) validateRedeemRequest(req *models.RedeemRequest) error {
	if req.TxHash == "" {
		return fmt.Errorf("txHash is required")
	}
	if req.NetworkChainID == "" {
		return fmt.Errorf("networkChainId is required")
	}
	if req.AmountTransfer == "" {
		return fmt.Errorf("amountTransfer is required")
	}
	if req.BankAccount == "" {
		return fmt.Errorf("bankAccount is required")
	}
	if req.BankCode == "" {
		return fmt.Errorf("bankCode is required")
	}
	if req.BankName == "" {
		return fmt.Errorf("bankName is required")
	}
	if req.BankAccountName == "" {
		return fmt.Errorf("bankAccountName is required")
	}
	if req.WalletAddress == "" {
		return fmt.Errorf("walletAddress is required")
	}
	if req.Notes == "" {
		return fmt.Errorf("notes is required")
	}
	return nil
}

// validateBridgeRequest validates the bridge request fields.
func (s *TransactionService) validateBridgeRequest(req *models.BridgeRequest) error {
	if req.FromChainID == "" {
		return fmt.Errorf("fromChainId is required")
	}
	if req.ToChainID == "" {
		return fmt.Errorf("toChainId is required")
	}
	if req.Amount == "" {
		return fmt.Errorf("amount is required")
	}
	if req.WalletAddress == "" {
		return fmt.Errorf("walletAddress is required")
	}
	if req.DestinationAddress == "" {
		return fmt.Errorf("destinationAddress is required")
	}
	return nil
}

// validateRatesRequest validates that exactly one amount is provided.
func (s *TransactionService) validateRatesRequest(req *models.RatesRequest) error {
	hasIDRXAmount := req.IDRXAmount != nil && *req.IDRXAmount != ""
	hasUSDTAmount := req.USDTAmount != nil && *req.USDTAmount != ""

	if !hasIDRXAmount && !hasUSDTAmount {
		return fmt.Errorf("either idrxAmount or usdtAmount must be provided")
	}

	if hasIDRXAmount && hasUSDTAmount {
		return fmt.Errorf("only one of idrxAmount or usdtAmount should be provided")
	}

	return nil
}

// validateTransactionHistoryRequest validates the transaction history request fields.
func (s *TransactionService) validateTransactionHistoryRequest(req *models.TransactionHistoryRequest) error {
	if req.TransactionType == "" {
		return fmt.Errorf("transactionType is required")
	}
	if req.Page <= 0 {
		return fmt.Errorf("page must be greater than 0")
	}
	if req.Take <= 0 {
		return fmt.Errorf("take must be greater than 0")
	}
	if req.Take > 100 {
		return fmt.Errorf("take cannot be greater than 100")
	}
	return nil
}
