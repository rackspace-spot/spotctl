package internal

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotctl/internal/ui"
)

const (
	kubernetesVersion1_31_1  = "1.31.1"
	kubernetesVersion1_30_10 = "1.30.10"
	kubernetesVersion1_29_6  = "1.29.6"
	cniCalico                = "calico"
	cniCilium                = "cilium"
	cniBringYourOwn          = "bring your own CNI"
)

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

	// Set default selection if provided - Not directly used in BubbleTea but keeping for fallback
	if defaultRegion != "" {
		for _, region := range regions {
			if region.Name == defaultRegion {
				// Default is handled by the UI component
				break
			}
		}
	}

	// Create and run the BubbleTea select prompt
	model := ui.NewSelectModel(regionOptions)
	p := tea.NewProgram(model)
	// Run the program and get the result
	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	// Get the selected option
	selectedModel, ok := m.(ui.SelectModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type: %T", m)
	}
	if selectedModel.Cancelled() {
		return "", context.Canceled
	}

	selectedOption := selectedModel.Selected()

	if selectedOption == "" {
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

	model := ui.NewSelectModel(serverClassOptions)
	p := tea.NewProgram(model)

	m, err := p.Run()
	if err != nil {
		return "", "", "", fmt.Errorf("error running prompt: %w", err)
	}

	selectedModel, ok := m.(ui.SelectModel)
	if !ok {
		return "", "", "", fmt.Errorf("unexpected model type: %T", m)
	}
	if selectedModel.Cancelled() {
		return "", "", "", context.Canceled
	}

	selectedOption := selectedModel.Selected()

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
		kubernetesVersion1_30_10,
		kubernetesVersion1_29_6,
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

	model := ui.NewSelectModel(versions)
	p := tea.NewProgram(model)

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	selectedModel, ok := m.(ui.SelectModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type: %T", m)
	}
	if selectedModel.Cancelled() {
		return "", context.Canceled
	}

	selected := selectedModel.Selected()

	if selected == "" && defaultVersion != "" {
		selected = defaultVersion
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

	model := ui.NewSelectModel(cniOptions)
	p := tea.NewProgram(model)

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	selectedModel, ok := m.(ui.SelectModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type: %T", m)
	}
	if selectedModel.Cancelled() {
		return "", context.Canceled
	}

	selected := selectedModel.Selected()

	if selected == "" && defaultCNI != "" {
		selected = defaultCNI
	}

	return selected, nil
}

// PromptForString prompts the user to enter a string value
func PromptForString(message, defaultValue string) (string, error) {
	model := ui.NewInputModel(message, defaultValue, false)
	p := tea.NewProgram(model)

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	inputModel, ok := m.(ui.InputModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type: %T", m)
	}
	if inputModel.Cancelled() {
		return "", context.Canceled
	}
	result := inputModel.Value()

	// If empty, return the default value
	if result == "" {
		return defaultValue, nil
	}

	return result, nil
}

// PromptForBidPrice prompts the user to enter a bid price for a spot node pool
func (c *Client) PromptForBidPrice(message, defaultValue string) (string, error) {
	if message == "" {
		message = "Enter your maximum bid price"
	}
	return PromptForString(message, defaultValue)
}

// Confirm prompts the user for a yes/no confirmation
func Confirm(message string, defaultYes bool) (bool, error) {
	model := ui.NewConfirmModel(message, defaultYes)
	p := tea.NewProgram(model)

	m, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("error running confirmation: %w", err)
	}

	confirmModel, ok := m.(ui.ConfirmModel)
	if !ok {
		return false, fmt.Errorf("unexpected model type: %T", m)
	}
	return confirmModel.Result(), nil
}

// PromptForNodeCount prompts the user to enter the number of nodes for a node pool
func (c *Client) PromptForNodeCount(poolType string) (string, error) {
	defaultNodes := "1"
	if poolType == "" {
		poolType = "node"
	}

	promptMessage := fmt.Sprintf("Enter number of %s nodes (default: 1)", poolType)

	// Run the input prompt
	model := ui.NewInputModel(promptMessage, defaultNodes, false)
	p := tea.NewProgram(model)

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	inputModel, ok := m.(ui.InputModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type: %T", m)
	}
	if inputModel.Cancelled() {
		return "", context.Canceled
	}

	// Get the result from the input model
	result := inputModel.Value()

	// Validate the input
	if result == "" {
		result = defaultNodes
	}

	return result, nil
}

// PromptForPoolType prompts the user to select a node pool type (Spot or On-Demand)
func (c *Client) PromptForPoolType() (string, error) {
	poolTypes := []string{"Spot", "On-Demand"}

	model := ui.NewSelectModel(poolTypes)
	p := tea.NewProgram(model)

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	selectedModel, ok := m.(ui.SelectModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type: %T", m)
	}
	if selectedModel.Cancelled() {
		return "", context.Canceled
	}

	return selectedModel.Selected(), nil
}

// GetOnDemandPrice retrieves the on-demand price for a given region and server class
func (c *Client) GetOnDemandPrice(region, serverClass string) (string, error) {
	serverClassList, err := c.api.ListServerClasses(context.Background(), region)
	if err != nil {
		return "", fmt.Errorf("failed to list server classes for region %s: %w", region, err)
	}

	for _, sc := range serverClassList.Items {
		if sc.Name == serverClass {
			return sc.OnDemandPricePerHour, nil
		}
	}

	return "", fmt.Errorf("could not find on-demand price for server class %s in region %s", serverClass, region)
}
