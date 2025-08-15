package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	config "github.com/rackspace-spot/spotcli/pkg"
)

// Client wraps the Spot SDK client with CLI-specific functionality
type Client struct {
	api rxtspot.SpotAPI
}

// NewClient creates a new CLI client
func NewClient() (*Client, error) {

	spotConfig, _ := config.LoadConfig()
	if spotConfig.Token == "" {
		return nil, fmt.Errorf("token is required. Please run 'spotctl configure' to set it up")
	}

	cfg := rxtspot.Config{
		BaseURL:      "https://spot.rackspace.com",
		OAuthURL:     "https://login.spot.rackspace.com",
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
		RefreshToken: spotConfig.Token,
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
