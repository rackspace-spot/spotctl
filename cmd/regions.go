package cmd

import (
	"context"
	"fmt"

	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"
)

var regionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "Manage regions",
	Long:  `Manage Rackspace Spot regions.`,
}

var regionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List regions",
	Long:  `List all regions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		regions, err := client.GetAPI().ListRegions(context.Background())
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		return internal.OutputData(regions, outputFormat)
	},
}

var regionsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get region",
	Long:  `Get a specific region.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("name is required")
		}
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		regions, err := client.GetAPI().GetRegion(context.Background(), name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		return internal.OutputData(regions, outputFormat)
	},
}

func init() {
	rootCmd.AddCommand(regionsCmd)
	regionsCmd.AddCommand(regionsListCmd)
	regionsCmd.AddCommand(regionsGetCmd)

	regionsGetCmd.Flags().String("name", "", "Region name")
	regionsListCmd.Flags().StringP("output", "o", "json", "Output format (json, table, yaml)")
}
