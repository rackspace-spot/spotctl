package cmd

import (
	"context"
	"fmt"

	"github.com/rackspace-spot/spotcli/internal"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Rackspace Spot",
	Long: `Authenticate with Rackspace Spot using your refresh token.
	
Set your refresh token as an environment variable:
export SPOT_REFRESH_TOKEN=your_refresh_token_here`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Validate authentication by listing organizations
		orgs, err := client.GetAPI().ListOrganizations(context.Background())
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		fmt.Println("Authentication successful!")
		if verbose {
			fmt.Printf("Authenticated organizations: %v\n", orgs)
		}
		//fmt.Printf("client token - %+v \n", client.Authenticate(context.Background()))
		return nil
	},
}

func init() {
	// Add authCmd as a subcommand
	rootCmd.AddCommand(authCmd)
}
