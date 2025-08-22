package cmd

import (
	"context"
	"fmt"

	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"
)

var serverclassesCmd = &cobra.Command{
	Use:   "serverclasses",
	Short: "Manage serverclasses",
	Long:  `Manage Rackspace Spot serverclasses.`,
}

var serverclassesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List serverclasses",
	Long:  `List all serverclasses.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		region, _ := cmd.Flags().GetString("region")

		serverclasses, err := client.GetAPI().ListServerClasses(context.Background(), region)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(serverclasses, outputFormat)
	},
}

var serverclassesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get serverclass",
	Long:  `Get a specific serverclass.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		serverclasses, err := client.GetAPI().GetServerClass(context.Background(), name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		return internal.OutputData(serverclasses, outputFormat)
	},
}

func init() {
	rootCmd.AddCommand(serverclassesCmd)
	serverclassesCmd.AddCommand(serverclassesListCmd)
	serverclassesCmd.AddCommand(serverclassesGetCmd)

	serverclassesGetCmd.Flags().String("name", "", "Serverclass name")
	serverclassesGetCmd.MarkFlagRequired("name")

	serverclassesListCmd.Flags().StringP("region", "r", "", "Region name")
	serverclassesListCmd.Flags().StringP("output", "o", "json", "Output format (json, table, yaml)")
}
