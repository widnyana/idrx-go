package idrx

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/igun997/idrx-go/models"
)

// AccountService handles account management operations including user onboarding,
// member management, and bank account operations.
type AccountService struct {
	client *Client
}

// Onboard creates a new user account with KYC data and ID document.
// This operation requires business-level authentication.
func (s *AccountService) Onboard(ctx context.Context, req *models.OnboardingRequest) (*models.OnboardingResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("onboarding request cannot be nil")
	}

	// Validate required fields
	if err := s.validateOnboardingRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create multipart form data
	body, contentType, err := s.createOnboardingMultipartForm(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart form: %w", err)
	}

	// Create request with custom content type for multipart
	path := "/api/auth/onboarding"
	fullURL := s.client.baseURL + path

	httpReq, err := s.client.buildMultipartRequestWithBody(ctx, "POST", fullURL, body, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Apply authentication with the multipart form data for signature
	if err := s.client.auth.Authenticate(ctx, httpReq, req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Send request manually since it's multipart
	resp, err := s.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the close error with context for debugging
			fmt.Printf("WARN [IDRX-SDK] failed to close response body for onboarding request (status: %d): %v\n", resp.StatusCode, closeErr)
		}
	}()

	// Parse response
	var response models.APIResponse[models.OnboardingResponse]
	if err := s.client.parseResponse(resp, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// GetMembers retrieves the list of members registered under the organization.
// This operation requires business-level authentication.
func (s *AccountService) GetMembers(ctx context.Context) ([]models.Member, error) {
	var response models.MembersResponse

	if err := s.client.doRequest(ctx, "GET", "/api/auth/members", nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	return response.Data, nil
}

// AddBankAccount adds a new bank account and creates an associated deposit address.
// This operation requires user-level authentication.
func (s *AccountService) AddBankAccount(ctx context.Context, req *models.AddBankAccountRequest) (*models.BankAccount, error) {
	if req == nil {
		return nil, fmt.Errorf("add bank account request cannot be nil")
	}

	// Validate required fields
	if req.BankAccountNumber == "" {
		return nil, fmt.Errorf("bank account number is required")
	}
	if req.BankCode == "" {
		return nil, fmt.Errorf("bank code is required")
	}

	// Create multipart form data
	body, contentType, err := s.createBankAccountMultipartForm(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart form: %w", err)
	}

	// Build and send multipart request
	path := "/api/auth/add-bank-account"
	fullURL := s.client.baseURL + path

	httpReq, err := s.client.buildMultipartRequestWithBody(ctx, "POST", fullURL, body, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Apply authentication
	if err := s.client.auth.Authenticate(ctx, httpReq, req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Send request
	resp, err := s.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the close error with context for debugging
			fmt.Printf("WARN [IDRX-SDK] failed to close response body for add bank account request (status: %d): %v\n", resp.StatusCode, closeErr)
		}
	}()

	// Parse response
	var response models.BankAccountResponse
	if err := s.client.parseResponse(resp, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// GetBankAccounts retrieves the user's registered bank accounts.
// This operation requires user-level authentication.
func (s *AccountService) GetBankAccounts(ctx context.Context) ([]models.BankAccount, error) {
	var response models.BankAccountsResponse

	if err := s.client.doRequest(ctx, "GET", "/api/auth/get-bank-accounts", nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}

	return response.Data, nil
}

// DeleteBankAccount deletes a specific bank account by ID.
// This operation requires user-level authentication.
func (s *AccountService) DeleteBankAccount(ctx context.Context, bankID int) error {
	path := fmt.Sprintf("/api/auth/delete-bank-account/%d", bankID)

	var response models.APIResponse[models.DeleteResponse]
	if err := s.client.doRequest(ctx, "DELETE", path, nil, &response); err != nil {
		return fmt.Errorf("failed to delete bank account: %w", err)
	}

	return nil
}

// validateOnboardingRequest validates the onboarding request fields.
func (s *AccountService) validateOnboardingRequest(req *models.OnboardingRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Fullname == "" {
		return fmt.Errorf("fullname is required")
	}
	if req.Address == "" {
		return fmt.Errorf("address is required")
	}
	if req.IDNumber == "" {
		return fmt.Errorf("ID number is required")
	}
	if req.IDFile == nil {
		return fmt.Errorf("ID file is required")
	}
	return nil
}

// createOnboardingMultipartForm creates multipart form data for onboarding request.
func (s *AccountService) createOnboardingMultipartForm(req *models.OnboardingRequest) (*bytes.Buffer, string, error) {
	// Prepare fields
	fields := map[string]string{
		"email":    req.Email,
		"fullname": req.Fullname,
		"address":  req.Address,
		"idNumber": req.IDNumber,
	}

	// Prepare files
	files := make(map[string]*os.File)
	if req.IDFile != nil {
		// Reset file pointer before use
		if _, err := req.IDFile.Seek(0, 0); err != nil {
			return nil, "", fmt.Errorf("failed to reset file pointer: %w", err)
		}
		files["idFile"] = req.IDFile
	}

	// Use generic multipart form builder
	body, writer, err := s.client.createMultipartForm(context.Background(), fields, files)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create multipart form: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

// createBankAccountMultipartForm creates multipart form data for bank account request.
func (s *AccountService) createBankAccountMultipartForm(req *models.AddBankAccountRequest) (*bytes.Buffer, string, error) {
	// Prepare fields
	fields := map[string]string{
		"bankAccountNumber": req.BankAccountNumber,
		"bankCode":          req.BankCode,
	}

	// No files for bank account request
	files := make(map[string]*os.File)

	// Use generic multipart form builder
	body, writer, err := s.client.createMultipartForm(context.Background(), fields, files)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create multipart form: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}
