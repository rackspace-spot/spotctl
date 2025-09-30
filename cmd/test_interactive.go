// +build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create a client with tokens from config
	client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// Test region selection
	region, err := client.PromptForRegion(context.Background())
	if err != nil {
		fmt.Printf("Error selecting region: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Selected region: %s\n", region)

	// Test server class selection
	serverClass, minBid, onDemandPrice, err := client.PromptForServerClassWithBidPrice(context.Background(), region, "spot")
	if err != nil {
		fmt.Printf("Error selecting server class: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Selected server class: %s (Min bid: %s, On-demand: %s)\n", serverClass, minBid, onDemandPrice)

	// Test Kubernetes version selection
	k8sVersion, err := client.PromptForKubernetesVersion("1.31.1")
	if err != nil {
		fmt.Printf("Error selecting Kubernetes version: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Selected Kubernetes version: %s\n", k8sVersion)

	// Test CNI selection
	cni, err := client.PromptForCNI("calico")
	if err != nil {
		fmt.Printf("Error selecting CNI: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Selected CNI: %s\n", cni)

	// Test node count
	nodeCount, err := client.PromptForNodeCount("spot")
	if err != nil {
		fmt.Printf("Error getting node count: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Entered node count: %s\n", nodeCount)

	// Test confirmation
	confirm, err := internal.Confirm("Would you like to proceed?", true)
	if err != nil {
		fmt.Printf("Error getting confirmation: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Confirmation: %v\n", confirm)
}
