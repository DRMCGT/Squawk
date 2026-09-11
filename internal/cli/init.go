package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squawk-dev/squawk/internal/adb"
	"github.com/squawk-dev/squawk/internal/session"
)

func newInitCmd() *cobra.Command {
	var sessionDir, device string
	var logLines int

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Start a new Squawk session",
		Long: "Start a new Squawk session and make it the active session.\n" +
			"Verifies adb and picks a connected device before creating the session.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(cmd.Context(), sessionDir, device, logLines)
		},
	}
	cmd.Flags().StringVar(&sessionDir, "session-dir", "", "root directory for Squawk data (default ./.squawk)")
	cmd.Flags().StringVar(&device, "device", "", "ADB device serial to test")
	cmd.Flags().IntVar(&logLines, "log-lines", 200, "number of logcat lines to capture per squawk")
	return cmd
}

func runInit(ctx context.Context, sessionDir, device string, logLines int) error {
	store, err := session.NewStore(sessionDir)
	if err != nil {
		return err
	}
	client := adb.NewClient(adb.NewExecRunner())
	if err := client.Check(ctx); err != nil {
		return err
	}
	devices, err := client.Devices(ctx)
	if err != nil {
		return err
	}
	selected, err := adb.SelectDevice(devices, device)
	if err != nil {
		return err
	}
	sess, err := store.CreateSession(selected.Serial, logLines)
	if err != nil {
		return err
	}
	fmt.Printf("squawk: session %s started on device %s\n", sess.ID, selected.Serial)
	fmt.Println("squawk: run `squawk watch` or `squawk capture --note \"...\"` to capture bugs")
	fmt.Println("squawk: run `squawk report` to generate a report")
	return nil
}
