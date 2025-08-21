package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rackspace-spot/spotcli/internal"
	config "github.com/rackspace-spot/spotcli/pkg"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set up Spot CLI defaults",
	Long:  `configure default orgID, token, and region for the Spot CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Organization ID: ")
		orgID, _ := reader.ReadString('\n')
		orgID = strings.TrimSpace(orgID)

		fmt.Print("Refresh Token: ")
		refreshToken, _ := reader.ReadString('\n')
		refreshToken = strings.TrimSpace(refreshToken)

		fmt.Print("Preferred Region: ")
		region, _ := reader.ReadString('\n')
		region = strings.TrimSpace(region)

		client, err := internal.NewClientWithTokens(refreshToken, "")
		if err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		access_token, err := client.Authenticate(context.Background())
		if err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		cfg := &config.SpotConfig{
			Org:          orgID,
			RefreshToken: refreshToken,
			AccessToken:  access_token,
			Region:       region,
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println("Configuration saved to ~/.spot_config")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
