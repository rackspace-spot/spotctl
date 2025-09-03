package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set up Spot CLI defaults",
	Long:  `configure default orgID, token, and region for the Spot CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Organization ID: ")
		orgID, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read organization ID: %w", err)
		}
		orgID = strings.TrimSpace(orgID)

		fmt.Print("Refresh Token: ")
		refreshToken, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read refresh token: %w", err)
		}
		refreshToken = strings.TrimSpace(refreshToken)

		fmt.Print("Preferred Region: ")
		region, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read preferred region: %w", err)
		}
		region = strings.TrimSpace(region)

		if region == "" {
			return fmt.Errorf("region is required")
		}
		if !isValidRegion(region) {
			return fmt.Errorf("region %s is not valid. Available regions: %s, %s, %s, %s, %s, %s, %s, %s", region, US_CENTRAL_ORD_1, HKG_HKG_1, AUS_SYD_1, UK_LON_1, US_EAST_IAD_1, US_CENTRAL_DFW_1, US_CENTRAL_DFW_2, US_WEST_SJC_1)
		}

		client, err := internal.NewClientWithTokens(refreshToken, "")
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		access_token, err := client.Authenticate(context.Background())
		if err != nil {
			return fmt.Errorf("%w", err)
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
