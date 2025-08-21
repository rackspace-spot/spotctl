package internal

// NewClientWithTokens is a convenience function to create a new client with just tokens
func NewClientWithTokens(refreshToken, accessToken string) (*Client, error) {
	cfg := ClientConfig{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}
	// Apply defaults
	defaultCfg := DefaultConfig()
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultCfg.BaseURL
	}
	if cfg.OAuthURL == "" {
		cfg.OAuthURL = defaultCfg.OAuthURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultCfg.Timeout
	}

	return NewClient(cfg)
}
