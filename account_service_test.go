package idrx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/igun997/idrx-go/models"
)

func TestAccountService_GetMembers(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/auth/members") {
			t.Errorf("expected members endpoint, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{
			"status": "ok",
			"data": [
				{
					"id": 1,
					"email": "test@example.com",
					"fullname": "Test User",
					"createdAt": "2024-01-01T00:00:00Z"
				}
			]
		}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewBusinessAuth("test-key", testSecretKey()),
	}

	service := &AccountService{client: client}

	ctx := context.Background()
	members, err := service.GetMembers(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}

	member := members[0]
	if member.ID != 1 {
		t.Errorf("expected member ID 1, got %d", member.ID)
	}
	if member.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", member.Email)
	}
	if member.Fullname != "Test User" {
		t.Errorf("expected fullname Test User, got %s", member.Fullname)
	}
}

func TestAccountService_GetBankAccounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/auth/get-bank-accounts") {
			t.Errorf("expected bank accounts endpoint, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{
			"status": "ok",
			"data": [
				{
					"id": 1,
					"bankAccountNumber": "123456789",
					"bankCode": "BCA",
					"bankName": "Bank Central Asia",
					"isActive": true,
					"createdAt": "2024-01-01T00:00:00Z",
					"updatedAt": "2024-01-01T00:00:00Z",
					"DepositWalletAddress": {
						"id": 1,
						"address": "0x123...",
						"chainId": "137",
						"chainName": "Polygon",
						"tokenAddress": "0x456...",
						"tokenSymbol": "USDC",
						"isActive": true,
						"createdAt": "2024-01-01T00:00:00Z",
						"updatedAt": "2024-01-01T00:00:00Z"
					}
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

	service := &AccountService{client: client}

	ctx := context.Background()
	accounts, err := service.GetBankAccounts(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}

	account := accounts[0]
	if account.ID != 1 {
		t.Errorf("expected account ID 1, got %d", account.ID)
	}
	if account.BankAccountNumber != "123456789" {
		t.Errorf("expected account number 123456789, got %s", account.BankAccountNumber)
	}
	if account.BankCode != "BCA" {
		t.Errorf("expected bank code BCA, got %s", account.BankCode)
	}

	// Verify deposit wallet
	if account.DepositWalletAddress.ChainID != "137" {
		t.Errorf("expected chain ID 137, got %s", account.DepositWalletAddress.ChainID)
	}
	if account.DepositWalletAddress.TokenSymbol != "USDC" {
		t.Errorf("expected token symbol USDC, got %s", account.DepositWalletAddress.TokenSymbol)
	}
}

func TestAccountService_DeleteBankAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/auth/delete-bank-account/123") {
			t.Errorf("expected delete endpoint with ID 123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{
			"status": "ok",
			"data": {
				"message": "Bank account deleted successfully"
			}
		}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &AccountService{client: client}

	ctx := context.Background()
	err := service.DeleteBankAccount(ctx, 123)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAccountService_OnboardValidation(t *testing.T) {
	client := &Client{
		baseURL:    "https://idrx.co",
		httpClient: &http.Client{},
		auth:       NewBusinessAuth("test-key", testSecretKey()),
	}

	service := &AccountService{client: client}
	ctx := context.Background()

	tests := []struct {
		name    string
		request *models.OnboardingRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "onboarding request cannot be nil",
		},
		{
			name: "missing email",
			request: &models.OnboardingRequest{
				Fullname: "Test User",
				Address:  "Test Address",
				IDNumber: "123456",
				IDFile:   &os.File{},
			},
			wantErr: "email is required",
		},
		{
			name: "missing fullname",
			request: &models.OnboardingRequest{
				Email:    "test@example.com",
				Address:  "Test Address",
				IDNumber: "123456",
				IDFile:   &os.File{},
			},
			wantErr: "fullname is required",
		},
		{
			name: "missing address",
			request: &models.OnboardingRequest{
				Email:    "test@example.com",
				Fullname: "Test User",
				IDNumber: "123456",
				IDFile:   &os.File{},
			},
			wantErr: "address is required",
		},
		{
			name: "missing ID number",
			request: &models.OnboardingRequest{
				Email:    "test@example.com",
				Fullname: "Test User",
				Address:  "Test Address",
				IDFile:   &os.File{},
			},
			wantErr: "ID number is required",
		},
		{
			name: "missing ID file",
			request: &models.OnboardingRequest{
				Email:    "test@example.com",
				Fullname: "Test User",
				Address:  "Test Address",
				IDNumber: "123456",
			},
			wantErr: "ID file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Onboard(ctx, tt.request)
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

func TestAccountService_AddBankAccountValidation(t *testing.T) {
	client := &Client{
		baseURL:    "https://idrx.co",
		httpClient: &http.Client{},
		auth:       NewUserAuth("user-key", testSecretKey()),
	}

	service := &AccountService{client: client}
	ctx := context.Background()

	tests := []struct {
		name    string
		request *models.AddBankAccountRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "add bank account request cannot be nil",
		},
		{
			name: "missing bank account number",
			request: &models.AddBankAccountRequest{
				BankCode: "BCA",
			},
			wantErr: "bank account number is required",
		},
		{
			name: "missing bank code",
			request: &models.AddBankAccountRequest{
				BankAccountNumber: "123456789",
			},
			wantErr: "bank code is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.AddBankAccount(ctx, tt.request)
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

func TestAccountService_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		io.WriteString(w, `{
			"statusCode": 400,
			"code": "INVALID_REQUEST",
			"message": "Invalid request data"
		}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewBusinessAuth("test-key", testSecretKey()),
	}

	service := &AccountService{client: client}

	ctx := context.Background()
	_, err := service.GetMembers(ctx)

	if err == nil {
		t.Fatal("expected error for bad request")
	}

	// The error should contain information about the API error
	if !strings.Contains(err.Error(), "Invalid request data") {
		t.Errorf("expected error to contain API error message, got %v", err)
	}
}

func TestAccountService_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, `{"status": "ok", "data": []}`)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		auth:       NewBusinessAuth("test-key", testSecretKey()),
	}

	service := &AccountService{client: client}

	// Create context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GetMembers(ctx)

	if err == nil {
		t.Error("expected error with canceled context")
	}

	if !strings.Contains(err.Error(), "context") {
		t.Errorf("expected context-related error, got %v", err)
	}
}

func TestValidateOnboardingRequest(t *testing.T) {
	service := &AccountService{}

	validReq := &models.OnboardingRequest{
		Email:    "test@example.com",
		Fullname: "Test User",
		Address:  "Test Address",
		IDNumber: "123456",
		IDFile:   &os.File{},
	}

	err := service.validateOnboardingRequest(validReq)
	if err != nil {
		t.Errorf("expected no error for valid request, got %v", err)
	}

	// Test each required field
	invalidReqs := []struct {
		name string
		req  *models.OnboardingRequest
	}{
		{"empty email", &models.OnboardingRequest{Fullname: "Test", Address: "Test", IDNumber: "123", IDFile: &os.File{}}},
		{"empty fullname", &models.OnboardingRequest{Email: "test@example.com", Address: "Test", IDNumber: "123", IDFile: &os.File{}}},
		{"empty address", &models.OnboardingRequest{Email: "test@example.com", Fullname: "Test", IDNumber: "123", IDFile: &os.File{}}},
		{"empty ID number", &models.OnboardingRequest{Email: "test@example.com", Fullname: "Test", Address: "Test", IDFile: &os.File{}}},
		{"no ID file", &models.OnboardingRequest{Email: "test@example.com", Fullname: "Test", Address: "Test", IDNumber: "123"}},
	}

	for _, test := range invalidReqs {
		t.Run(test.name, func(t *testing.T) {
			err := service.validateOnboardingRequest(test.req)
			if err == nil {
				t.Errorf("expected error for %s", test.name)
			}
		})
	}
}
