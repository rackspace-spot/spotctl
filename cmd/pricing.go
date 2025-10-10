package cmd

import (
	"context"
	"fmt"

	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"
)

var pricingCmd = &cobra.Command{
	Use:   "pricing",
	Short: "Manage pricing",
	Long:  `Manage Rackspace Spot pricing.`,
}

var pricingGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get market price for a server class",
	Long:  `Get market price for a server class.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		serverclass, _ := cmd.Flags().GetString("serverclass")

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pricing, err := client.GetAPI().GetPriceDetailsForServerClass(context.Background(), serverclass)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		return internal.OutputData(pricing, outputFormat)
	},
}

var pricingGetAllServerClassCmd = &cobra.Command{
	Use:   "get-all",
	Short: "Get market price for all server classes",
	Long:  `Get market price for all server classes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		pricing, err := client.GetAPI().GetPriceDetails(context.Background())
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		return internal.OutputData(pricing, outputFormat)
	},
}

func init() {
	rootCmd.AddCommand(pricingCmd)
	pricingCmd.AddCommand(pricingGetCmd)
	pricingCmd.AddCommand(pricingGetAllServerClassCmd)
	pricingGetCmd.Flags().String("serverclass", "", "Serverclass name")
}
