package cmd

import (
	"context"
	"fmt"

	"github.com/rackspace-spot/spotcli/internal"
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

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		region, _ := cmd.Flags().GetString("region")

		serverclasses, err := client.GetAPI().ListServerClasses(context.Background(), region)
		if err != nil {
			return fmt.Errorf("failed to list serverclasses: %w", err)
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

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		serverclasses, err := client.GetAPI().GetServerClass(context.Background(), name)
		if err != nil {
			return fmt.Errorf("failed to list serverclasses: %w", err)
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
