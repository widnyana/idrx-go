package idrx

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/igun997/idrx-go/blockchain"
)

// BlockchainService provides blockchain-specific operations for IDRX
type BlockchainService struct {
	client *blockchain.Client
}

// NewBlockchainService creates a new blockchain service
func NewBlockchainService(client *blockchain.Client) *BlockchainService {
	return &BlockchainService{
		client: client,
	}
}

// GetBalance returns the IDRX balance for an address on the specified chain
func (bs *BlockchainService) GetBalance(ctx context.Context, chainID uint64, address string) (*blockchain.TokenAmount, error) {
	addr := common.HexToAddress(address)
	return bs.client.BalanceOf(ctx, chainID, addr)
}

// GetBalanceByNetwork returns the IDRX balance for an address on the specified network
func (bs *BlockchainService) GetBalanceByNetwork(ctx context.Context, networkName string, address string) (*blockchain.TokenAmount, error) {
	networkConfig, exists := blockchain.GetNetworkConfig(networkName)
	if !exists {
		return nil, fmt.Errorf("network %s not supported", networkName)
	}
	return bs.GetBalance(ctx, networkConfig.ChainID, address)
}

// GetTokenInfo returns basic token information for the specified chain
func (bs *BlockchainService) GetTokenInfo(ctx context.Context, chainID uint64) (*blockchain.TokenInfo, error) {
	return bs.client.GetTokenInfo(ctx, chainID)
}

// GetTokenInfoByNetwork returns basic token information for the specified network
func (bs *BlockchainService) GetTokenInfoByNetwork(ctx context.Context, networkName string) (*blockchain.TokenInfo, error) {
	networkConfig, exists := blockchain.GetNetworkConfig(networkName)
	if !exists {
		return nil, fmt.Errorf("network %s not supported", networkName)
	}
	return bs.GetTokenInfo(ctx, networkConfig.ChainID)
}

// Transfer transfers IDRX tokens between addresses
func (bs *BlockchainService) Transfer(ctx context.Context, chainID uint64, toAddress string, amount string) (*TransferResult, error) {
	to := common.HexToAddress(toAddress)
	tokenAmount, err := blockchain.ParseTokenAmount(amount, 18)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	tx, err := bs.client.Transfer(ctx, chainID, to, tokenAmount)
	if err != nil {
		return nil, err
	}

	return &TransferResult{
		TxHash:  tx.Hash().Hex(),
		ChainID: chainID,
		To:      toAddress,
		Amount:  amount,
		Status:  "pending",
	}, nil
}

// TransferResult represents the result of a transfer operation
type TransferResult struct {
	TxHash  string `json:"txHash"`
	ChainID uint64 `json:"chainId"`
	To      string `json:"to"`
	Amount  string `json:"amount"`
	Status  string `json:"status"`
}

// BurnForRedemption burns IDRX tokens for fiat redemption
func (bs *BlockchainService) BurnForRedemption(
	ctx context.Context,
	chainID uint64,
	amount string,
	accountNumber string,
) (*BurnResult, error) {
	tokenAmount, err := blockchain.ParseTokenAmount(amount, 18)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	tx, err := bs.client.BurnWithAccountNumber(ctx, chainID, tokenAmount, accountNumber)
	if err != nil {
		return nil, err
	}

	return &BurnResult{
		TxHash:        tx.Hash().Hex(),
		ChainID:       chainID,
		Amount:        amount,
		AccountNumber: accountNumber,
		Status:        "pending",
	}, nil
}

// BurnResult represents the result of a burn operation
type BurnResult struct {
	TxHash        string `json:"txHash"`
	ChainID       uint64 `json:"chainId"`
	Amount        string `json:"amount"`
	AccountNumber string `json:"accountNumber"`
	Status        string `json:"status"`
}

// InitiateBridge initiates a cross-chain bridge transaction
func (bs *BlockchainService) InitiateBridge(ctx context.Context, request *BridgeTransactionRequest) (*BridgeResult, error) {
	if !blockchain.IsChainSupported(request.FromChainID) {
		return nil, fmt.Errorf("source chain %d not supported", request.FromChainID)
	}

	if !blockchain.IsChainSupported(request.ToChainID) {
		return nil, fmt.Errorf("destination chain %d not supported", request.ToChainID)
	}

	tokenAmount, err := blockchain.ParseTokenAmount(request.Amount, 18)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	bridgeRequest := &blockchain.BridgeRequest{
		Amount:      tokenAmount,
		ToChainID:   request.ToChainID,
		ToAddress:   common.HexToAddress(request.ToAddress),
		FromChainID: request.FromChainID,
	}

	tx, err := bs.client.BurnBridge(ctx, bridgeRequest)
	if err != nil {
		return nil, err
	}

	return &BridgeResult{
		TxHash:      tx.Hash().Hex(),
		FromChainID: request.FromChainID,
		ToChainID:   request.ToChainID,
		ToAddress:   request.ToAddress,
		Amount:      request.Amount,
		Status:      "burning",
	}, nil
}

// BridgeTransactionRequest represents a bridge transaction request
type BridgeTransactionRequest struct {
	FromChainID uint64 `json:"fromChainId" validate:"required"`
	ToChainID   uint64 `json:"toChainId" validate:"required"`
	ToAddress   string `json:"toAddress" validate:"required"`
	Amount      string `json:"amount" validate:"required"`
}

// BridgeResult represents the result of a bridge operation
type BridgeResult struct {
	TxHash      string `json:"txHash"`
	FromChainID uint64 `json:"fromChainId"`
	ToChainID   uint64 `json:"toChainId"`
	ToAddress   string `json:"toAddress"`
	Amount      string `json:"amount"`
	Status      string `json:"status"`
}

// GetTransactionStatus returns the status of a blockchain transaction
func (bs *BlockchainService) GetTransactionStatus(ctx context.Context, chainID uint64, txHash string) (*TransactionStatus, error) {
	client, err := bs.client.GetClient(chainID)
	if err != nil {
		return nil, err
	}

	hash := common.HexToHash(txHash)
	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil {
		// Transaction might still be pending
		_, pending, err := client.TransactionByHash(ctx, hash)
		if err != nil {
			return nil, fmt.Errorf("failed to check transaction status: %w", err)
		}

		if pending {
			return &TransactionStatus{
				TxHash:  txHash,
				ChainID: chainID,
				Status:  "pending",
			}, nil
		}
	}

	status := "success"
	if receipt.Status != types.ReceiptStatusSuccessful {
		status = "failed"
	}

	return &TransactionStatus{
		TxHash:      txHash,
		ChainID:     chainID,
		Status:      status,
		BlockNumber: receipt.BlockNumber.Uint64(),
		GasUsed:     receipt.GasUsed,
	}, nil
}

// TransactionStatus represents the status of a blockchain transaction
type TransactionStatus struct {
	TxHash      string `json:"txHash"`
	ChainID     uint64 `json:"chainId"`
	Status      string `json:"status"` // "pending", "success", "failed", "not_found"
	BlockNumber uint64 `json:"blockNumber,omitempty"`
	GasUsed     uint64 `json:"gasUsed,omitempty"`
}

// GetNetworkInfo returns information about supported networks
func (bs *BlockchainService) GetNetworkInfo() []NetworkInfo {
	networks := make([]NetworkInfo, 0, len(blockchain.SupportedNetworks))

	for networkName, config := range blockchain.SupportedNetworks {
		networks = append(networks, NetworkInfo{
			NetworkName:     networkName,
			ChainID:         config.ChainID,
			Name:            config.Name,
			BlockTime:       config.BlockTime,
			ContractAddress: config.ContractAddress.Hex(),
			IsTestnet:       config.IsTestnet,
		})
	}

	return networks
}

// NetworkInfo represents information about a supported network
type NetworkInfo struct {
	NetworkName     string        `json:"networkName"`
	ChainID         uint64        `json:"chainId"`
	Name            string        `json:"name"`
	BlockTime       time.Duration `json:"blockTime"`
	ContractAddress string        `json:"contractAddress"`
	IsTestnet       bool          `json:"isTestnet"`
}

// WaitForTransaction waits for a transaction to be confirmed
func (bs *BlockchainService) WaitForTransaction(ctx context.Context, chainID uint64, txHash string) (*TransactionStatus, error) {
	hash := common.HexToHash(txHash)
	receipt, err := bs.client.WaitForTransaction(ctx, chainID, hash)
	if err != nil {
		return nil, err
	}

	status := "success"
	if receipt.Status != types.ReceiptStatusSuccessful {
		status = "failed"
	}

	return &TransactionStatus{
		TxHash:      txHash,
		ChainID:     chainID,
		Status:      status,
		BlockNumber: receipt.BlockNumber.Uint64(),
		GasUsed:     receipt.GasUsed,
	}, nil
}

// GetPlatformFees returns the current platform fee configuration
func (bs *BlockchainService) GetPlatformFees(ctx context.Context, chainID uint64) (*PlatformFees, error) {
	feeInfo, err := bs.client.GetPlatformFeeInfo(ctx, chainID)
	if err != nil {
		return nil, err
	}

	return &PlatformFees{
		ChainID:       chainID,
		Recipient:     feeInfo.Recipient.Hex(),
		BurnBridgeFee: feeInfo.BurnBridgeFee,
		MintBridgeFee: feeInfo.MintBridgeFee,
	}, nil
}

// PlatformFees represents platform fee information
type PlatformFees struct {
	ChainID       uint64 `json:"chainId"`
	Recipient     string `json:"recipient"`
	BurnBridgeFee uint64 `json:"burnBridgeFee"`
	MintBridgeFee uint64 `json:"mintBridgeFee"`
}

// IsAddressBlacklisted checks if an address is blacklisted
func (bs *BlockchainService) IsAddressBlacklisted(ctx context.Context, chainID uint64, address string) (bool, error) {
	addr := common.HexToAddress(address)
	return bs.client.IsBlacklisted(ctx, chainID, addr)
}

// GetBridgeNonce returns the current bridge nonce for a chain
func (bs *BlockchainService) GetBridgeNonce(ctx context.Context, chainID uint64) (string, error) {
	nonce, err := bs.client.GetBridgeNonce(ctx, chainID)
	if err != nil {
		return "", err
	}
	return nonce.String(), nil
}
