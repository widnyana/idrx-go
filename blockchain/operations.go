// Package blockchain provides blockchain client operations and token management for IDRX tokens.
package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

// TokenAmount represents a precise token amount using decimal
type TokenAmount struct {
	Amount   decimal.Decimal
	Decimals int32
}

// ToWei converts the token amount to wei (smallest unit)
func (ta *TokenAmount) ToWei() *big.Int {
	multiplier := decimal.NewFromInt(10).Pow(decimal.NewFromInt32(ta.Decimals))
	wei := ta.Amount.Mul(multiplier)
	return wei.BigInt()
}

// FromWei creates a TokenAmount from wei
func FromWei(wei *big.Int, decimals int32) *TokenAmount {
	divisor := decimal.NewFromInt(10).Pow(decimal.NewFromInt32(decimals))
	amount := decimal.NewFromBigInt(wei, 0).Div(divisor)
	return &TokenAmount{
		Amount:   amount,
		Decimals: decimals,
	}
}

// BalanceOf returns the IDRX balance for an address on the specified chain
func (c *Client) BalanceOf(_ context.Context, chainID uint64, address common.Address) (*TokenAmount, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	balance, err := contract.BalanceOf(nil, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return FromWei(balance, int32(GetDecimals(chainID))), nil
}

// TotalSupply returns the total supply of IDRX on the specified chain
func (c *Client) TotalSupply(_ context.Context, chainID uint64) (*TokenAmount, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	supply, err := contract.TotalSupply(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total supply: %w", err)
	}

	return FromWei(supply, int32(GetDecimals(chainID))), nil
}

// Transfer sends IDRX tokens from the client's address to another address
func (c *Client) Transfer(ctx context.Context, chainID uint64, to common.Address, amount *TokenAmount) (*types.Transaction, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	transactor, err := c.CreateTransactor(ctx, chainID)
	if err != nil {
		return nil, err
	}

	tx, err := contract.Transfer(transactor, to, amount.ToWei())
	if err != nil {
		return nil, fmt.Errorf("failed to transfer tokens: %w", err)
	}

	return tx, nil
}

// Mint mints new IDRX tokens to the specified address (requires MINTER_ROLE)
func (c *Client) Mint(ctx context.Context, chainID uint64, to common.Address, amount *TokenAmount) (*types.Transaction, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	transactor, err := c.CreateTransactor(ctx, chainID)
	if err != nil {
		return nil, err
	}

	tx, err := contract.Mint(transactor, to, amount.ToWei())
	if err != nil {
		return nil, fmt.Errorf("failed to mint tokens: %w", err)
	}

	return tx, nil
}

// Burn burns IDRX tokens from the client's address
func (c *Client) Burn(ctx context.Context, chainID uint64, amount *TokenAmount) (*types.Transaction, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	transactor, err := c.CreateTransactor(ctx, chainID)
	if err != nil {
		return nil, err
	}

	tx, err := contract.Burn(transactor, amount.ToWei())
	if err != nil {
		return nil, fmt.Errorf("failed to burn tokens: %w", err)
	}

	return tx, nil
}

// BurnWithAccountNumber burns IDRX tokens for fiat redemption with account number
func (c *Client) BurnWithAccountNumber(
	ctx context.Context,
	chainID uint64,
	amount *TokenAmount,
	accountNumber string,
) (*types.Transaction, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	transactor, err := c.CreateTransactor(ctx, chainID)
	if err != nil {
		return nil, err
	}

	tx, err := contract.BurnWithAccountNumber(transactor, amount.ToWei(), accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to burn tokens with account number: %w", err)
	}

	return tx, nil
}

// BridgeRequest represents a cross-chain bridge request
type BridgeRequest struct {
	Amount      *TokenAmount
	ToChainID   uint64
	ToAddress   common.Address
	FromChainID uint64
}

// BurnBridge burns tokens on the source chain for bridging to another chain
func (c *Client) BurnBridge(ctx context.Context, request *BridgeRequest) (*types.Transaction, error) {
	contract, err := c.GetContract(request.FromChainID)
	if err != nil {
		return nil, err
	}

	transactor, err := c.CreateTransactor(ctx, request.FromChainID)
	if err != nil {
		return nil, err
	}

	tx, err := contract.BurnBridge(transactor, request.Amount.ToWei(), new(big.Int).SetUint64(request.ToChainID))
	if err != nil {
		return nil, fmt.Errorf("failed to burn for bridge: %w", err)
	}

	return tx, nil
}

// MintBridge mints tokens on the destination chain after bridge burn (requires MINTER_ROLE)
func (c *Client) MintBridge(
	ctx context.Context,
	chainID uint64,
	to common.Address,
	amount *TokenAmount,
	fromChainID uint64,
	bridgeNonce *big.Int,
) (*types.Transaction, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	transactor, err := c.CreateTransactor(ctx, chainID)
	if err != nil {
		return nil, err
	}

	tx, err := contract.MintBridge(transactor, to, amount.ToWei(), new(big.Int).SetUint64(fromChainID), bridgeNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to mint bridge tokens: %w", err)
	}

	return tx, nil
}

// GetBridgeNonce returns the current bridge nonce for the specified chain
func (c *Client) GetBridgeNonce(_ context.Context, chainID uint64) (*big.Int, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	nonce, err := contract.BridgeNonce(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get bridge nonce: %w", err)
	}

	return nonce, nil
}

// IsNonceUsed checks if a bridge nonce has been used for a specific chain
func (c *Client) IsNonceUsed(_ context.Context, chainID uint64, fromChainID uint64, nonce *big.Int) (bool, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return false, err
	}

	used, err := contract.FromChainNonceUsed(nil, new(big.Int).SetUint64(fromChainID), nonce)
	if err != nil {
		return false, fmt.Errorf("failed to check nonce usage: %w", err)
	}

	return used, nil
}

// PlatformFeeInfo represents platform fee information
type PlatformFeeInfo struct {
	Recipient     common.Address
	BurnBridgeFee uint64
	MintBridgeFee uint64
}

// GetPlatformFeeInfo returns the current platform fee configuration
func (c *Client) GetPlatformFeeInfo(_ context.Context, chainID uint64) (*PlatformFeeInfo, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	recipient, burnFee, mintFee, err := contract.GetPlatformFeeInfo(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform fee info: %w", err)
	}

	return &PlatformFeeInfo{
		Recipient:     recipient,
		BurnBridgeFee: burnFee,
		MintBridgeFee: mintFee,
	}, nil
}

// IsBlacklisted checks if an address is blacklisted
func (c *Client) IsBlacklisted(_ context.Context, chainID uint64, address common.Address) (bool, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return false, err
	}

	blacklisted, err := contract.GetBlackListStatus(nil, address)
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist status: %w", err)
	}

	return blacklisted, nil
}

// TokenInfo represents basic token information
type TokenInfo struct {
	Name     string
	Symbol   string
	Decimals uint8
}

// GetTokenInfo returns basic information about the IDRX token
func (c *Client) GetTokenInfo(_ context.Context, chainID uint64) (*TokenInfo, error) {
	contract, err := c.GetContract(chainID)
	if err != nil {
		return nil, err
	}

	name, err := contract.Name(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get token name: %w", err)
	}

	symbol, err := contract.Symbol(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get token symbol: %w", err)
	}

	decimals, err := contract.Decimals(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get token decimals: %w", err)
	}

	return &TokenInfo{
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}, nil
}

// ParseTokenAmount creates a TokenAmount from a string representation
func ParseTokenAmount(amountStr string, decimals int32) (*TokenAmount, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	return &TokenAmount{
		Amount:   amount,
		Decimals: decimals,
	}, nil
}

// String returns a human-readable string representation of the TokenAmount
func (ta *TokenAmount) String() string {
	return ta.Amount.StringFixed(ta.Decimals)
}

// CompleteBridge performs a complete bridge operation from source to destination chain
func (c *Client) CompleteBridge(ctx context.Context, request *BridgeRequest) error {
	// Step 1: Burn tokens on source chain
	burnTx, err := c.BurnBridge(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to burn on source chain: %w", err)
	}

	// Step 2: Wait for burn transaction to be mined
	receipt, err := c.WaitForTransaction(ctx, request.FromChainID, burnTx.Hash())
	if err != nil {
		return fmt.Errorf("burn transaction failed: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("burn transaction reverted")
	}

	// Step 3: Extract bridge nonce from burn event
	bridgeNonce, err := c.extractBridgeNonceFromReceipt(receipt)
	if err != nil {
		return fmt.Errorf("failed to extract bridge nonce: %w", err)
	}

	// Step 4: Mint tokens on destination chain (requires MINTER_ROLE)
	_, err = c.MintBridge(ctx, request.ToChainID, request.ToAddress, request.Amount, request.FromChainID, bridgeNonce)
	if err != nil {
		return fmt.Errorf("failed to mint on destination chain: %w", err)
	}

	return nil
}

// extractBridgeNonceFromReceipt extracts the bridge nonce from a burn bridge transaction receipt
func (c *Client) extractBridgeNonceFromReceipt(receipt *types.Receipt) (*big.Int, error) {
	// Parse BurnBridge event logs to extract the nonce
	for _, log := range receipt.Logs {
		if len(log.Topics) > 0 {
			// BurnBridge event signature: BurnBridge(address,uint256,uint256,uint256,uint256,uint256)
			burnBridgeTopic := common.HexToHash("0x") // This would need to be the actual event signature hash
			if log.Topics[0] == burnBridgeTopic {
				// Extract nonce from log data (this is simplified - actual implementation would need proper ABI decoding)
				// For now, return a placeholder
				return big.NewInt(0), nil
			}
		}
	}

	return nil, fmt.Errorf("bridge nonce not found in transaction receipt")
}
