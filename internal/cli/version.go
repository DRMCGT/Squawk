package cli

import (
	"fmt"

	"github.com/DRMCGT/Squawk/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build metadata",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println(version.String())
			return nil
		},
	}
}
