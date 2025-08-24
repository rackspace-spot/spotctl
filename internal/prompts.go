package internal

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

const kubernetesVersion1_31_1 = "1.31.1"
const kubernetesVersion1_30_0 = "1.30.0"
const kubernetesVersion1_29_0 = "1.29.0"
const cniCalico = "calico"
const cniCilium = "cilium"
const cniBringYourOwn = "bring your own CNI"

// PromptForRegion prompts the user to select a region from the available regions
func (c *Client) PromptForRegion(ctx context.Context) (string, error) {
	return c.PromptForRegionWithDefault(ctx, "")
}

// PromptForRegionWithDefault prompts the user to select a region with an optional default using a dropdown
func (c *Client) PromptForRegionWithDefault(ctx context.Context, defaultRegion string) (string, error) {
	regions, err := c.api.ListRegions(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list available regions: %w", err)
	}

	if len(regions) == 0 {
		return "", fmt.Errorf("no regions available")
	}

	// Sort regions by name for consistent display
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].Name < regions[j].Name
	})

	// Prepare options for the dropdown
	var regionOptions []string
	regionMap := make(map[string]string) // maps display string to region name

	for _, region := range regions {
		display := region.Name
		if region.Description != "" {
			display = fmt.Sprintf("%s - %s", region.Name, region.Description)
		}
		regionOptions = append(regionOptions, display)
		regionMap[display] = region.Name
	}

	// Set default selection if provided
	var defaultOption string
	if defaultRegion != "" {
		for _, region := range regions {
			if region.Name == defaultRegion {
				defaultOption = region.Name
				if region.Description != "" {
					defaultOption = fmt.Sprintf("%s - %s", region.Name, region.Description)
				}
				break
			}
		}
	}

	// Create and configure the prompt
	var selectedOption string
	prompt := &survey.Select{
		Message: "Select a region:",
		Options: regionOptions,
		Default: defaultOption,
	}

	// Show the prompt
	if err := survey.AskOne(prompt, &selectedOption); err != nil {
		// Fallback to the old method if there's an error with the dropdown
		return c.fallbackRegionPrompt(regions, defaultRegion)
	}

	// Return the actual region name
	return regionMap[selectedOption], nil
}

// fallbackRegionPrompt provides a fallback method for region selection if the dropdown fails
func (c *Client) fallbackRegionPrompt(regions []rxtspot.Region, defaultRegion string) (string, error) {
	// Find default region index if provided
	defaultIndex := -1
	if defaultRegion != "" {
		for i, r := range regions {
			if r.Name == defaultRegion {
				defaultIndex = i
				break
			}
		}
	}

	// Display available regions
	fmt.Println("\nAvailable regions:")
	for i, region := range regions {
		desc := region.Name
		if region.Description != "" {
			desc = fmt.Sprintf("%s - %s", region.Name, region.Description)
		}
		prefix := "  "
		if i == defaultIndex {
			prefix = "* "
		}
		fmt.Printf("%s%d. %s\n", prefix, i+1, desc)
	}

	// Simple input prompt
	var selectedIndex int
	for {
		prompt := "\nEnter the number of the region"
		if defaultIndex >= 0 {
			prompt = fmt.Sprintf("%s [%d]: ", prompt, defaultIndex+1)
		} else {
			prompt = fmt.Sprintf("%s: ", prompt)
		}

		fmt.Print(prompt)
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil && err.Error() != "unexpected newline" {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		// If user just pressed Enter and there's a default, use it
		if input == "" && defaultIndex >= 0 {
			return regions[defaultIndex].Name, nil
		}

		// Otherwise parse the input
		selectedIndex, err = strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		if selectedIndex < 1 || selectedIndex > len(regions) {
			fmt.Printf("Please enter a number between 1 and %d\n", len(regions))
			continue
		}

		break
	}

	return regions[selectedIndex-1].Name, nil
}

// PromptForServerClass prompts the user to select a server class for the given region
func (c *Client) PromptForServerClass(ctx context.Context, region string) (string, error) {
	serverClass, _, _, err := c.PromptForServerClassWithBidPrice(ctx, region, "")
	return serverClass, err
}

// PromptForServerClassWithBidPrice prompts the user to select a server class and returns the class name, minimum bid price, and on-demand price
// poolType should be either "spot" or "ondemand" to determine which pricing information to display
func (c *Client) PromptForServerClassWithBidPrice(ctx context.Context, region, poolType string) (string, string, string, error) {
	serverClassList, err := c.api.ListServerClasses(ctx, region)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to list server classes for region %s: %w", region, err)
	}

	if serverClassList == nil || len(serverClassList.Items) == 0 {
		return "", "", "", fmt.Errorf("no server classes available for region %s", region)
	}

	type serverClassInfo struct {
		Name                      string
		CPU                       string
		Memory                    string
		CurrentMarketPricePerHour string
		MinBidPricePerHour        string
		OnDemandPricePerHour      string
	}

	var serverClassOptions []string
	serverClassMap := make(map[string]serverClassInfo)

	for _, sc := range serverClassList.Items {
		if sc.MinBidPricePerHour < sc.CurrentMarketPricePerHour {
			sc.MinBidPricePerHour = sc.CurrentMarketPricePerHour
		}
		info := serverClassInfo{
			Name:                      sc.Name,
			CPU:                       sc.Resources.CPU,
			Memory:                    sc.Resources.Memory,
			CurrentMarketPricePerHour: sc.CurrentMarketPricePerHour,
			MinBidPricePerHour:        sc.MinBidPricePerHour,
			OnDemandPricePerHour:      sc.OnDemandPricePerHour,
		}
		var desc string
		// Handle both "on-demand" and "ondemand" for backward compatibility
		if poolType == "ondemand" || poolType == "on-demand" {
			desc = fmt.Sprintf("%s (CPU: %s, Memory: %s, Price: %s)",
				info.Name, info.CPU, info.Memory, info.OnDemandPricePerHour)
		} else {
			desc = fmt.Sprintf("%s (CPU: %s, Memory: %s, Current Market Price: %s, Min Bid Price: %s)",
				info.Name, info.CPU, info.Memory, info.CurrentMarketPricePerHour, info.MinBidPricePerHour)
		}
		serverClassOptions = append(serverClassOptions, desc)
		serverClassMap[desc] = info
	}

	var selectedOption string
	prompt := &survey.Select{
		Message: "Select a server class:",
		Options: serverClassOptions,
	}

	if err := survey.AskOne(prompt, &selectedOption); err != nil {
		return "", "", "", fmt.Errorf("failed to select server class: %w", err)
	}

	// Get the selected server class info
	info := serverClassMap[selectedOption]

	// Use the server class's current market price as the minimum bid price
	// Remove any non-numeric characters except decimal point
	minBidPriceStr := strings.TrimPrefix(info.MinBidPricePerHour, "$")
	minBidPriceStr = strings.TrimSpace(minBidPriceStr)

	OnDemandPricePerHour := strings.TrimPrefix(info.OnDemandPricePerHour, "$")
	OnDemandPricePerHour = strings.TrimSpace(OnDemandPricePerHour)

	return info.Name, minBidPriceStr, OnDemandPricePerHour, nil
}

// PromptForKubernetesVersion prompts the user to select a Kubernetes version
func (c *Client) PromptForKubernetesVersion(defaultVersion string) (string, error) {
	// These are common Kubernetes versions, you might want to fetch them from an API
	versions := []string{
		kubernetesVersion1_31_1,
		kubernetesVersion1_30_0,
		kubernetesVersion1_29_0,
	}

	// If default version is not in the list, add it
	versionExists := false
	for _, v := range versions {
		if v == defaultVersion {
			versionExists = true
			break
		}
	}

	if !versionExists && defaultVersion != "" {
		versions = append([]string{defaultVersion}, versions...)
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select Kubernetes version:",
		Options: versions,
		Default: defaultVersion,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", fmt.Errorf("failed to select Kubernetes version: %w", err)
	}

	return selected, nil
}

// PromptForCNI prompts the user to select a CNI plugin
func (c *Client) PromptForCNI(defaultCNI string) (string, error) {
	cniOptions := []string{
		cniCalico,
		cniCilium,
		cniBringYourOwn,
	}

	// If default CNI is not in the list, add it
	cniExists := false
	for _, cni := range cniOptions {
		if cni == defaultCNI {
			cniExists = true
			break
		}
	}

	if !cniExists && defaultCNI != "" {
		cniOptions = append([]string{defaultCNI}, cniOptions...)
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select CNI plugin:",
		Options: cniOptions,
		Default: defaultCNI,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", fmt.Errorf("failed to select CNI plugin: %w", err)
	}

	return selected, nil
}

// PromptForString prompts the user to enter a string value
func PromptForString(message, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return "", fmt.Errorf("failed to get input: %w", err)
	}

	return result, nil
}

// Confirm prompts the user for a yes/no confirmation
func Confirm(message string, defaultYes bool) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultYes,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}

	return result, nil
}

// PromptForNodeCount prompts the user to enter the number of nodes for a node pool
func (c *Client) PromptForNodeCount(poolType string) (string, error) {
	defaultNodes := "1"
	if poolType == "" {
		poolType = "node"
	}

	var result string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Enter number of %s nodes (default: 1):", poolType),
		Default: defaultNodes,
	}

	// Add validation
	validate := func(val interface{}) error {
		str, ok := val.(string)
		if !ok {
			return fmt.Errorf("invalid input type")
		}

		num, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("please enter a valid number")
		}

		if num < 1 {
			return fmt.Errorf("number of nodes must be at least 1")
		}

		return nil
	}

	if err := survey.AskOne(prompt, &result, survey.WithValidator(validate)); err != nil {
		return "", fmt.Errorf("failed to get node count: %w", err)
	}

	// If user just pressed enter, use the default
	if result == "" {
		return defaultNodes, nil
	}

	return result, nil
}
