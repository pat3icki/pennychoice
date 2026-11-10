package monnify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	LIVE_BASEURL    = "https://api.monnify.com"
	SANDBOX_BASEURL = "https://sandbox.monnify.com"
)

type contextKey string

const authRetryKey contextKey = "authRetryed"

type Client struct {
	baseUrl         string
	apiKey          string
	secretKey       string
	basicToken      string
	bearerToken     string
	expire          time.Time
	isAuthenticated bool

	http *http.Client
	send func(ctx context.Context, c *Client, method, url string, reqBody, respBody interface{}) error
}

// JSONFloat64 is a custom type to handle float64 fields that might be represented as numeric values or string values in JSON response.
type JSONFloat64 float64

// UnmarshalJSON implements json.Unmarshaler
func (f *JSONFloat64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*f = 0
		return nil
	}

	// Try unmarshaling as float64 first
	var val float64
	if err := json.Unmarshal(data, &val); err == nil {
		*f = JSONFloat64(val)
		return nil
	}

	// Try unmarshaling as string
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "" {
		*f = 0
		return nil
	}

	// Parse string to float64
	var parsedVal float64
	if _, err := fmt.Sscanf(str, "%f", &parsedVal); err != nil {
		return fmt.Errorf("failed to parse JSONFloat64 string %q: %w", str, err)
	}

	*f = JSONFloat64(parsedVal)
	return nil
}

// New is the monnify config initializer
func New(httpClient *http.Client, baseURL, apiKey, secretKey string) *Client {
	baseUrl := fmt.Sprintf("%s/api", baseURL)
	client := &Client{
		http:      httpClient,
		apiKey:    apiKey,
		secretKey: secretKey,
		baseUrl:   baseUrl,
	}
	client.send = defaultSend

	return client
}

// defaultSend is the default HTTP request handler for Monnify API
func defaultSend(ctx context.Context, c *Client, method, url string, reqBody, respBody interface{}) error {
	// Ensure authentication is valid
	if err := c.ensureAuthenticated(ctx); err != nil {
		return fmt.Errorf("authentication error: %w", err)
	}

	// Marshal request body if provided
	var reqReader io.Reader
	if reqBody != nil {
		reqData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqReader = bytes.NewReader(reqData)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, reqReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusUnauthorized {
			c.isAuthenticated = false
			// Check if we already retried
			if alreadyRetried, _ := ctx.Value(authRetryKey).(bool); !alreadyRetried {
				// Re-authenticate
				if err := c.Authenticate(ctx); err == nil {
					// Retry sending the request once
					retryCtx := context.WithValue(ctx, authRetryKey, true)
					return c.send(retryCtx, c, method, url, reqBody, respBody)
				}
			}
		}

		var errResp struct {
			RequestSuccessful bool   `json:"requestSuccessful"`
			ResponseMessage   string `json:"responseMessage"`
			ResponseCode      string `json:"responseCode"`
		}
		if len(respData) > 0 {
			_ = json.Unmarshal(respData, &errResp)
		}
		if errResp.ResponseMessage != "" {
			return fmt.Errorf("monnify API error (status %d, code %s): %s", resp.StatusCode, errResp.ResponseCode, errResp.ResponseMessage)
		}
		return fmt.Errorf("monnify API HTTP error: status %d, body %s", resp.StatusCode, string(respData))
	}

	// Unmarshal response body if provided
	if respBody != nil && len(respData) > 0 {
		if err := json.Unmarshal(respData, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Authenticate handles the authentication with Monnify API
func (c *Client) Authenticate(ctx context.Context) error {
	authURL := c.baseUrl + "/v1/auth/login"

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth basic request: %w", err)
	}
	c.basicToken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.apiKey, c.secretKey)))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.basicToken))

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		RequestSuccessful bool   `json:"requestSuccessful"`
		ResponseMessage   string `json:"responseMessage"`
		ResponseCode      string `json:"responseCode"`
		ResponseBody      struct {
			AccessToken string `json:"accessToken"`
			ExpiresIn   int    `json:"expiresIn"` // in seconds
		} `json:"responseBody"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	if !authResp.RequestSuccessful {
		return fmt.Errorf("authentication unsuccessful: %s", authResp.ResponseMessage)
	}

	c.bearerToken = authResp.ResponseBody.AccessToken
	c.expire = time.Now().Add(time.Duration(authResp.ResponseBody.ExpiresIn) * time.Second)
	c.isAuthenticated = true

	return nil
}

// ensureAuthenticated checks if authentication is needed and performs it
func (c *Client) ensureAuthenticated(ctx context.Context) error {
	// Check if token is expired or not set
	if !c.isAuthenticated || time.Now().After(c.expire) {
		return c.Authenticate(ctx)
	}
	return nil
}
