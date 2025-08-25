package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

// Client wraps the Spot SDK client with CLI-specific functionality
type Client struct {
	api rxtspot.SpotAPI
}

// ClientConfig holds configuration for creating a new Client
type ClientConfig struct {
	RefreshToken string
	AccessToken  string
	BaseURL      string
	OAuthURL     string
	Timeout      time.Duration
}

// DefaultConfig returns a default ClientConfig with sensible defaults
func DefaultConfig() ClientConfig {
	return ClientConfig{
		BaseURL:  "https://spot.rackspace.com",
		OAuthURL: "https://login.spot.rackspace.com",
		Timeout:  30 * time.Second,
	}
}

// NewClient creates a new CLI client with the given configuration
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required. Please run 'spotctl configure' to set it up")
	}

	sdkCfg := rxtspot.Config{
		BaseURL:      cfg.BaseURL,
		OAuthURL:     cfg.OAuthURL,
		HTTPClient:   &http.Client{Timeout: cfg.Timeout},
		RefreshToken: cfg.RefreshToken,
		AccessToken:  cfg.AccessToken,
		Timeout:      cfg.Timeout,
	}

	client := rxtspot.NewSpotClient(sdkCfg)

	// Let the SDK handle token validation and refresh
	_, err := client.Authenticate(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return &Client{
		api: client,
	}, nil
}

// GetAPI returns the underlying Spot API client
func (c *Client) GetAPI() rxtspot.SpotAPI {
	return c.api
}

// Authenticate performs authentication
func (c *Client) Authenticate(ctx context.Context) (string, error) {
	return c.api.Authenticate(ctx)
}
