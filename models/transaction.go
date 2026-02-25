package models

import (
	"time"
)

// CustomerDetail represents optional customer information for mint requests.
type CustomerDetail struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
}

// MintRequest represents the request to mint IDRX or other stablecoins.
type MintRequest struct {
	ToBeMinted               string          `json:"toBeMinted" validate:"required"`
	DestinationWalletAddress string          `json:"destinationWalletAddress" validate:"required"`
	NetworkChainID           string          `json:"networkChainId" validate:"required"`
	ExpiryPeriod             int             `json:"expiryPeriod" validate:"required"`
	RequestType              RequestType     `json:"requestType,omitempty"`
	ProductDetails           string          `json:"productDetails,omitempty"`
	CustomerDetail           *CustomerDetail `json:"customerDetail,omitempty"`
}

// MintResponse represents the response from a mint request.
type MintResponse struct {
	PaymentURL     string    `json:"paymentUrl"`
	AdjustedAmount string    `json:"adjustedAmount"`
	Fees           []Fee     `json:"fees,omitempty"`
	TransactionID  string    `json:"transactionId,omitempty"`
	ExpiresAt      time.Time `json:"expiresAt,omitempty"`
}

// RedeemRequest represents the request to redeem IDRX to bank account.
type RedeemRequest struct {
	TxHash          string `json:"txHash" validate:"required"`
	NetworkChainID  string `json:"networkChainId" validate:"required"`
	AmountTransfer  string `json:"amountTransfer" validate:"required"`
	BankAccount     string `json:"bankAccount" validate:"required"`
	BankCode        string `json:"bankCode" validate:"required"`
	BankName        string `json:"bankName" validate:"required"`
	BankAccountName string `json:"bankAccountName" validate:"required"`
	WalletAddress   string `json:"walletAddress" validate:"required"`
	Notes           string `json:"notes" validate:"required"`
}

// RedeemResponse represents the response from a redeem request.
type RedeemResponse struct {
	TransactionID     string            `json:"transactionId"`
	Status            TransactionStatus `json:"status"`
	ReferenceNumber   string            `json:"referenceNumber,omitempty"`
	ProcessingTime    string            `json:"processingTime,omitempty"`
	EstimatedComplete time.Time         `json:"estimatedComplete,omitempty"`
}

// RatesRequest represents the request to get swap rates.
type RatesRequest struct {
	IDRXAmount *string `url:"idrxAmount,omitempty"`
	USDTAmount *string `url:"usdtAmount,omitempty"`
	ChainID    *string `url:"chainId,omitempty"`
}

// RatesResponse represents the response with current swap rates.
type RatesResponse struct {
	Price     string `json:"price"`
	BuyAmount string `json:"buyAmount"`
	FromToken string `json:"fromToken"`
	ToToken   string `json:"toToken"`
	ChainID   string `json:"chainId"`
}

// FeesRequest represents the request to get additional fees.
type FeesRequest struct {
	FeeType FeeType `url:"feeType" validate:"required"`
	ChainID *string `url:"chainId,omitempty"`
}

// FeesResponse represents the response with fee information.
type FeesResponse = APIResponse[[]Fee]

// TransactionHistoryRequest represents the request for transaction history.
type TransactionHistoryRequest struct {
	TransactionType TransactionType    `url:"transactionType" validate:"required"`
	Page            int                `url:"page" validate:"required,min=1"`
	Take            int                `url:"take" validate:"required,min=1,max=100"`
	Status          *TransactionStatus `url:"status,omitempty"`
	ChainID         *string            `url:"chainId,omitempty"`
	StartDate       *time.Time         `url:"startDate,omitempty"`
	EndDate         *time.Time         `url:"endDate,omitempty"`
	MinAmount       *string            `url:"minAmount,omitempty"`
	MaxAmount       *string            `url:"maxAmount,omitempty"`
}

// Transaction represents a transaction record.
type Transaction struct {
	ID              string            `json:"id"`
	Type            TransactionType   `json:"type"`
	Status          TransactionStatus `json:"status"`
	Amount          string            `json:"amount"`
	Currency        string            `json:"currency"`
	ChainID         string            `json:"chainId"`
	ChainName       string            `json:"chainName"`
	TxHash          string            `json:"txHash,omitempty"`
	WalletAddress   string            `json:"walletAddress,omitempty"`
	BankAccount     string            `json:"bankAccount,omitempty"`
	BankName        string            `json:"bankName,omitempty"`
	ReferenceNumber string            `json:"referenceNumber,omitempty"`
	Fees            []Fee             `json:"fees,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	CompletedAt     *time.Time        `json:"completedAt,omitempty"`
}

// TransactionHistoryResponse represents the paginated transaction history response.
type TransactionHistoryResponse struct {
	Data     []Transaction `json:"data"`
	Metadata ListMetadata  `json:"metadata"`
}

// BridgeRequest represents a request for token bridging operations.
type BridgeRequest struct {
	FromChainID        string `json:"fromChainId" validate:"required"`
	ToChainID          string `json:"toChainId" validate:"required"`
	Amount             string `json:"amount" validate:"required"`
	WalletAddress      string `json:"walletAddress" validate:"required"`
	DestinationAddress string `json:"destinationAddress" validate:"required"`
}

// BridgeResponse represents the response from a bridge request.
type BridgeResponse struct {
	TransactionID  string    `json:"transactionId"`
	Status         string    `json:"status"`
	EstimatedTime  string    `json:"estimatedTime"`
	BridgeFee      string    `json:"bridgeFee"`
	ExchangeRate   string    `json:"exchangeRate"`
	ExpectedAmount string    `json:"expectedAmount"`
	CreatedAt      time.Time `json:"createdAt"`
}
