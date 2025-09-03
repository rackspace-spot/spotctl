package main

import (
	"github.com/rackspace-spot/spotctl/cmd"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func main() {
	cobra.EnableTraverseRunHooks = true
	// Execute the root command
	defer klog.Flush() // Flushes logs before exit

	cmd.Execute()
}
