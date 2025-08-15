package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

		fmt.Print("API Token: ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)

		fmt.Print("Preferred Region: ")
		region, _ := reader.ReadString('\n')
		region = strings.TrimSpace(region)

		cfg := &config.SpotConfig{
			Org:    orgID,
			Token:  token,
			Region: region,
		}

		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Println("Configuration saved to ~/.spot_config")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
