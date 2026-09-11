package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/DRMCGT/Squawk/internal/adb"
	"github.com/spf13/cobra"
)

func newDevicesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "devices",
		Short: "List connected Android devices",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDevices(cmd.Context())
		},
	}
}

func runDevices(ctx context.Context) error {
	client := adb.NewClient(adb.NewExecRunner())
	devices, err := client.Devices(ctx)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		fmt.Println("no Android devices are connected")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DEVICE\tSTATE")
	for _, d := range devices {
		fmt.Fprintf(w, "%s\t%s\n", d.Serial, d.State)
	}
	return w.Flush()
}
