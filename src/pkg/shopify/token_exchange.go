package shopify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mbbipul/tryonmagic-backend/src/config"
)

type TokenExchangeRequest struct {
	ClientID           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
	GrantType          string `json:"grant_type"`
	SubjectToken       string `json:"subject_token"`
	SubjectTokenType   string `json:"subject_token_type"`
	RequestedTokenType string `json:"requested_token_type"`
}

type TokenExchangeResponse struct {
	AccessToken         string         `json:"access_token"`
	Scope               string         `json:"scope"`
	ExpiresIn           int            `json:"expires_in"`
	AssociatedUserScope string         `json:"associated_user_scope"`
	AssociatedUser      AssociatedUser `json:"associated_user"`
}

func TokenExchange(shop string, subjectToken string) (*Session, error) {
	url := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)
	fmt.Println("Token exchange URL:", url)
	payload := TokenExchangeRequest{
		ClientID:           config.ShopifyAPIKey,
		ClientSecret:       config.ShopifyAPISecret,
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		SubjectToken:       subjectToken,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:id_token",
		RequestedTokenType: "urn:shopify:params:oauth:token-type:online-access-token",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create a context with timeout to avoid hanging requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP client with proper timeouts and no connection reuse
	// to avoid context conflicts with fasthttp
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives:     true,
			DisableCompression:    false,
			MaxIdleConns:          0,
			MaxIdleConnsPerHost:   0,
			MaxConnsPerHost:       0,
			IdleConnTimeout:       0,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResponse TokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		return nil, errors.New("missing access token in response")
	}

	return CreateSession(&tokenResponse, shop)
}
