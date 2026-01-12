package idrx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/igun997/idrx-go/models"
)

// doRequest is the shared HTTP request method using the template method pattern.
// It handles request building, authentication, sending, and response parsing.
func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	if c.auth == nil {
		return fmt.Errorf("authentication provider not configured")
	}

	// Build full URL
	fullURL := c.baseURL + path

	// Create HTTP request
	req, err := c.buildHTTPRequest(ctx, method, fullURL, body)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// Apply authentication
	if err := c.auth.Authenticate(ctx, req, body); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the close error with context for debugging
			fmt.Printf("WARN [IDRX-SDK] failed to close response body for %s %s (status: %d): %v\n",
				req.Method, req.URL.Path, resp.StatusCode, closeErr)
		}
	}()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		return c.parseErrorResponse(resp.StatusCode, respBody)
	}

	// Parse success response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// buildHTTPRequest creates an HTTP request with the appropriate content type and body.
func (c *Client) buildHTTPRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	switch v := body.(type) {
	case *multipart.Writer:
		// Handle multipart form data
		return c.buildMultipartRequest(ctx, method, url, v)
	case nil:
		// No body
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	default:
		// JSON body
		return c.buildJSONRequest(ctx, method, url, v)
	}
}

// buildJSONRequest creates an HTTP request with JSON body.
func (c *Client) buildJSONRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// buildMultipartRequest creates an HTTP request with multipart form data.
func (c *Client) buildMultipartRequest(ctx context.Context, method, url string, body *multipart.Writer) (*http.Request, error) {
	// This is a placeholder - the actual multipart data should be prepared by the caller
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", body.FormDataContentType())
	return req, nil
}

// createMultipartForm creates a multipart form with the given fields and files.
// It respects context cancellation for long-running operations.
func (c *Client) createMultipartForm(
	ctx context.Context,
	fields map[string]string,
	files map[string]*os.File,
) (*bytes.Buffer, *multipart.Writer, error) {
	// Check if context is already canceled
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context canceled before creating multipart form: %w", err)
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, value := range fields {
		// Check context before each field operation
		if err := ctx.Err(); err != nil {
			return nil, nil, fmt.Errorf("context canceled during field writing: %w", err)
		}

		if err := writer.WriteField(key, value); err != nil {
			return nil, nil, fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}

	// Add files
	for key, file := range files {
		// Check context before each file operation
		if err := ctx.Err(); err != nil {
			return nil, nil, fmt.Errorf("context canceled during file processing: %w", err)
		}

		// Detect MIME type from file extension
		mimeType := c.detectMimeType(file.Name())

		// Create form part with proper MIME type
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, key, filepath.Base(file.Name())))
		h.Set("Content-Type", mimeType)

		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create form file %s: %w", key, err)
		}

		// Check context before potentially long file copy operation
		if err := ctx.Err(); err != nil {
			return nil, nil, fmt.Errorf("context canceled before copying file %s: %w", key, err)
		}

		if _, err := io.Copy(part, file); err != nil {
			return nil, nil, fmt.Errorf("failed to copy file %s: %w", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return body, writer, nil
}

// detectMimeType detects MIME type from file extension
func (c *Client) detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}

// parseErrorResponse parses an error response from the API.
func (c *Client) parseErrorResponse(statusCode int, body []byte) error {
	var apiErr models.APIError

	// Try to parse as structured error response
	if err := json.Unmarshal(body, &apiErr); err != nil {
		// If parsing fails, create a generic error
		return &models.APIError{
			StatusCode: statusCode,
			Message:    string(body),
		}
	}

	// Ensure status code is set
	if apiErr.StatusCode == 0 {
		apiErr.StatusCode = statusCode
	}

	return &apiErr
}

// buildQueryParams converts a struct to URL query parameters using reflection.
func (c *Client) buildQueryParams(params any) (string, error) {
	if params == nil {
		return "", nil
	}

	values := url.Values{}
	v := reflect.ValueOf(params)
	t := reflect.TypeOf(params)

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// Only process struct types
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("params must be a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get the URL tag
		urlTag := fieldType.Tag.Get("url")
		if urlTag == "" || urlTag == "-" {
			continue
		}

		// Parse tag options (e.g., "omitempty")
		tagName, tagOptions := parseTag(urlTag)

		// Skip empty values if omitempty is set
		if slices.Contains(tagOptions, "omitempty") && isEmptyValue(field) {
			continue
		}

		// Convert field value to string
		strValue := convertToString(field)
		if strValue != "" {
			values.Add(tagName, strValue)
		}
	}

	return values.Encode(), nil
}

// parseTag parses struct tag into name and options.
func parseTag(tag string) (string, []string) {
	parts := strings.Split(tag, ",")
	name := parts[0]
	options := parts[1:]
	return name, options
}

// isEmptyValue checks if a reflect.Value is empty.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Struct, reflect.UnsafePointer:
		return false
	}
	return false
}

// convertToString converts a reflect.Value to its string representation.
func convertToString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Ptr:
		if v.IsNil() {
			return ""
		}
		return convertToString(v.Elem())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Slice, reflect.Struct, reflect.UnsafePointer:
		return fmt.Sprintf("%v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// buildMultipartRequestWithBody creates an HTTP request with pre-built multipart body.
func (c *Client) buildMultipartRequestWithBody(
	ctx context.Context,
	method, url string,
	body *bytes.Buffer,
	contentType string,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	return req, nil
}

// parseResponse parses an HTTP response into the result struct.
func (c *Client) parseResponse(resp *http.Response, result any) error {
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		return c.parseErrorResponse(resp.StatusCode, respBody)
	}

	// Parse success response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
