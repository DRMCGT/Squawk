// Package cli wires the Squawk commands together: flag parsing, user-facing
// errors, and calls into the application services.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the root command and exits with a non-zero status on error.
func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "squawk:", err)
		os.Exit(1)
	}
}

// NewRootCmd builds the root command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "squawk",
		Short:         "Capture bug context (screenshot, logs, notes) during manual Android QA",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newInitCmd(),
		newCaptureCmd(),
		newWatchCmd(),
		newReportCmd(),
		newDevicesCmd(),
		newVersionCmd(),
	)
	return root
}
