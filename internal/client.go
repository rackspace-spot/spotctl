package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

// Client wraps the Spot SDK client with CLI-specific functionality
type Client struct {
	api rxtspot.SpotAPI
}

// NewClient creates a new CLI client
func NewClient() (*Client, error) {
	refreshToken := os.Getenv("SPOT_REFRESH_TOKEN")
	if refreshToken == "" {
		return nil, fmt.Errorf("SPOT_REFRESH_TOKEN environment variable is required")
	}
	cfg := rxtspot.Config{BaseURL: "https://spot.rackspace.com",
		OAuthURL:     "https://login.spot.rackspace.com",
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
		RefreshToken: refreshToken,
		Timeout:      30 * time.Second}

	client := rxtspot.NewClient(cfg)

	if err := client.Authenticate(context.Background()); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return &Client{api: client}, nil
}

// GetAPI returns the underlying Spot API client
func (c *Client) GetAPI() rxtspot.SpotAPI {
	return c.api
}

// Authenticate performs authentication
func (c *Client) Authenticate(ctx context.Context) error {
	return c.api.Authenticate(ctx)
}

// func (c *Client) ListSpotNodePools(ctx context.Context, org string, cloudspace string) ([]*rxtspot.SpotNodePool, error) {
// 	return c.api.ListSpotNodePools(ctx, org, cloudspace)
// }
