package cmd

import (
	"context"
	"fmt"

	"github.com/rackspace-spot/spotcli/internal"
	config "github.com/rackspace-spot/spotcli/pkg"
	"github.com/spf13/cobra"
)

var pricingCmd = &cobra.Command{
	Use:   "pricing",
	Short: "Manage pricing",
	Long:  `Manage Rackspace Spot pricing.`,
}

var pricingGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get pricing",
	Long:  `Get a specific pricing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		serverclass, _ := cmd.Flags().GetString("serverclass")

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		pricing, err := client.GetAPI().GetMarketPriceForServerClass(context.Background(), serverclass)
		if err != nil {
			return fmt.Errorf("failed to get pricing: %w", err)
		}
		return internal.OutputData(pricing, outputFormat)
	},
}

func init() {
	rootCmd.AddCommand(pricingCmd)
	pricingCmd.AddCommand(pricingGetCmd)
	pricingGetCmd.Flags().String("serverclass", "", "Serverclass name")
}
