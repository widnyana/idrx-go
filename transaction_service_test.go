package idrx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/igun997/idrx-go/models"
)

func TestTransactionService_MintRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/transaction/mint-request") {
			t.Errorf("expected mint request endpoint, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{
			"status": "ok",
			"data": {
				"paymentUrl": "https://example.com/pay",
				"adjustedAmount": "1000",
				"transactionId": "tx123",
				"expiresAt": "2024-01-01T12:00:00Z"
			}
		}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}

	ctx := context.Background()
	req := &models.MintRequest{
		ToBeMinted:               "1000",
		DestinationWalletAddress: "0x123...",
		NetworkChainID:           "137",
		ExpiryPeriod:             3600,
	}

	result, err := service.MintRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.PaymentURL != "https://example.com/pay" {
		t.Errorf("expected payment URL https://example.com/pay, got %s", result.PaymentURL)
	}
	if result.AdjustedAmount != "1000" {
		t.Errorf("expected adjusted amount 1000, got %s", result.AdjustedAmount)
	}
	if result.TransactionID != "tx123" {
		t.Errorf("expected transaction ID tx123, got %s", result.TransactionID)
	}
}

func TestTransactionService_RedeemRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/transaction/redeem-request") {
			t.Errorf("expected redeem request endpoint, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{
			"status": "ok",
			"data": {
				"transactionId": "redeem456",
				"status": "processing",
				"referenceNumber": "REF789",
				"processingTime": "2-3 business days"
			}
		}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}

	ctx := context.Background()
	req := &models.RedeemRequest{
		TxHash:          "0xabc123",
		NetworkChainID:  "137",
		AmountTransfer:  "500",
		BankAccount:     "1234567890",
		BankCode:        "BCA",
		BankName:        "Bank Central Asia",
		BankAccountName: "John Doe",
		WalletAddress:   "0x456...",
		Notes:           "Test redemption",
	}

	result, err := service.RedeemRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.TransactionID != "redeem456" {
		t.Errorf("expected transaction ID redeem456, got %s", result.TransactionID)
	}
	if string(result.Status) != "processing" {
		t.Errorf("expected status processing, got %s", string(result.Status))
	}
	if result.ReferenceNumber != "REF789" {
		t.Errorf("expected reference number REF789, got %s", result.ReferenceNumber)
	}
}

func TestTransactionService_GetRates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/transaction/rates") {
			t.Errorf("expected rates endpoint, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{
			"status": "ok",
			"data": {
				"price": "1.0",
				"buyAmount": "1000",
				"fromToken": "USDT",
				"toToken": "IDRX",
				"chainId": "137"
			}
		}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}

	ctx := context.Background()
	idrxAmount := "1000"
	req := &models.RatesRequest{
		IDRXAmount: &idrxAmount,
	}

	result, err := service.GetRates(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.BuyAmount != "1000" {
		t.Errorf("expected buy amount 1000, got %s", result.BuyAmount)
	}
	if result.Price != "1.0" {
		t.Errorf("expected price 1.0, got %s", result.Price)
	}
	if result.FromToken != "USDT" {
		t.Errorf("expected from token USDT, got %s", result.FromToken)
	}
}

func TestTransactionService_GetMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/transaction/method") {
			t.Errorf("expected method endpoint, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{
			"status": "ok",
			"data": [
				{
					"bankCode": "BCA",
					"bankName": "Bank Central Asia",
					"maxAmountTransfer": "100000000"
				}
			]
		}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}

	ctx := context.Background()
	banks, err := service.GetMethods(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(banks) != 1 {
		t.Fatalf("expected 1 bank, got %d", len(banks))
	}

	bank := banks[0]
	if bank.BankCode != "BCA" {
		t.Errorf("expected bank code BCA, got %s", bank.BankCode)
	}
	if bank.BankName != "Bank Central Asia" {
		t.Errorf("expected bank name Bank Central Asia, got %s", bank.BankName)
	}
}

func TestTransactionService_GetTransactionHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/transaction/user-transaction-history") {
			t.Errorf("expected transaction history endpoint, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{
			"status": "ok",
			"data": [
				{
					"id": "tx001",
					"type": "mint",
					"amount": "1000",
					"status": "completed",
					"currency": "IDRX",
					"chainId": "137",
					"createdAt": "2024-01-01T00:00:00Z",
					"updatedAt": "2024-01-01T01:00:00Z"
				}
			],
			"metadata": {
				"page": 1,
				"take": 10,
				"totalItems": 1,
				"totalPages": 1
			}
		}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}

	ctx := context.Background()
	req := &models.TransactionHistoryRequest{
		TransactionType: "mint",
		Page:            1,
		Take:            10,
	}

	result, err := service.GetTransactionHistory(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(result.Data))
	}

	tx := result.Data[0]
	if tx.ID != "tx001" {
		t.Errorf("expected transaction ID tx001, got %s", tx.ID)
	}
	if string(tx.Type) != "mint" {
		t.Errorf("expected type mint, got %s", string(tx.Type))
	}
	if tx.Amount != "1000" {
		t.Errorf("expected amount 1000, got %s", tx.Amount)
	}
}

func TestTransactionService_MintRequestValidation(t *testing.T) {
	client := &Client{
		baseURL:    "https://idrx.co",
		httpClient: &http.Client{},
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}
	ctx := context.Background()

	tests := []struct {
		name    string
		request *models.MintRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "mint request cannot be nil",
		},
		{
			name: "missing toBeMinted",
			request: &models.MintRequest{
				DestinationWalletAddress: "0x123",
				NetworkChainID:           "137",
				ExpiryPeriod:             3600,
			},
			wantErr: "toBeMinted is required",
		},
		{
			name: "missing destination wallet",
			request: &models.MintRequest{
				ToBeMinted:     "1000",
				NetworkChainID: "137",
				ExpiryPeriod:   3600,
			},
			wantErr: "destinationWalletAddress is required",
		},
		{
			name: "missing network chain ID",
			request: &models.MintRequest{
				ToBeMinted:               "1000",
				DestinationWalletAddress: "0x123",
				ExpiryPeriod:             3600,
			},
			wantErr: "networkChainId is required",
		},
		{
			name: "invalid expiry period",
			request: &models.MintRequest{
				ToBeMinted:               "1000",
				DestinationWalletAddress: "0x123",
				NetworkChainID:           "137",
				ExpiryPeriod:             0,
			},
			wantErr: "expiryPeriod must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.MintRequest(ctx, tt.request)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestTransactionService_RedeemRequestValidation(t *testing.T) {
	client := &Client{
		baseURL:    "https://idrx.co",
		httpClient: &http.Client{},
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}
	ctx := context.Background()

	tests := []struct {
		name    string
		request *models.RedeemRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "redeem request cannot be nil",
		},
		{
			name: "missing txHash",
			request: &models.RedeemRequest{
				NetworkChainID:  "137",
				AmountTransfer:  "500",
				BankAccount:     "123",
				BankCode:        "BCA",
				BankName:        "Bank",
				BankAccountName: "John",
				WalletAddress:   "0x456",
				Notes:           "Test",
			},
			wantErr: "txHash is required",
		},
		{
			name: "missing notes",
			request: &models.RedeemRequest{
				TxHash:          "0xabc",
				NetworkChainID:  "137",
				AmountTransfer:  "500",
				BankAccount:     "123",
				BankCode:        "BCA",
				BankName:        "Bank",
				BankAccountName: "John",
				WalletAddress:   "0x456",
			},
			wantErr: "notes is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.RedeemRequest(ctx, tt.request)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestTransactionService_RatesRequestValidation(t *testing.T) {
	client := &Client{
		baseURL:    "https://idrx.co",
		httpClient: &http.Client{},
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}
	ctx := context.Background()

	tests := []struct {
		name    string
		request *models.RatesRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "rates request cannot be nil",
		},
		{
			name:    "missing both amounts",
			request: &models.RatesRequest{},
			wantErr: "either idrxAmount or usdtAmount must be provided",
		},
		{
			name: "both amounts provided",
			request: &models.RatesRequest{
				IDRXAmount: stringPtr("1000"),
				USDTAmount: stringPtr("1000"),
			},
			wantErr: "only one of idrxAmount or usdtAmount should be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetRates(ctx, tt.request)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestTransactionService_TransactionHistoryValidation(t *testing.T) {
	client := &Client{
		baseURL:    "https://idrx.co",
		httpClient: &http.Client{},
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &TransactionService{client: client}
	ctx := context.Background()

	tests := []struct {
		name    string
		request *models.TransactionHistoryRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "transaction history request cannot be nil",
		},
		{
			name: "missing transaction type",
			request: &models.TransactionHistoryRequest{
				Page: 1,
				Take: 10,
			},
			wantErr: "transactionType is required",
		},
		{
			name: "invalid page",
			request: &models.TransactionHistoryRequest{
				TransactionType: "mint",
				Page:            0,
				Take:            10,
			},
			wantErr: "page must be greater than 0",
		},
		{
			name: "invalid take",
			request: &models.TransactionHistoryRequest{
				TransactionType: "mint",
				Page:            1,
				Take:            0,
			},
			wantErr: "take must be greater than 0",
		},
		{
			name: "take too large",
			request: &models.TransactionHistoryRequest{
				TransactionType: "mint",
				Page:            1,
				Take:            101,
			},
			wantErr: "take cannot be greater than 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetTransactionHistory(ctx, tt.request)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
