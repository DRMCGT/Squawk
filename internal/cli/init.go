package cli

import (
	"context"
	"fmt"

	"github.com/DRMCGT/Squawk/internal/session"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var sessionDir, backend, device, tab, cdp string
	var logLines int

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Start a new Squawk session",
		Long: "Start a new Squawk session and make it the active session.\n" +
			"Verifies the backend (adb by default, or Chrome DevTools with --backend browser)\n" +
			"and picks a target device or tab before creating the session.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			printBanner()
			return runInit(cmd.Context(), sessionDir, backend, device, tab, cdp, logLines)
		},
	}
	cmd.Flags().StringVar(&sessionDir, "session-dir", "", "root directory for Squawk data (default ./.squawk)")
	cmd.Flags().StringVar(&backend, "backend", "android", "capture backend: android or browser")
	cmd.Flags().StringVar(&device, "device", "", "ADB device serial to test")
	cmd.Flags().StringVar(&tab, "tab", "", "browser tab URL substring to test")
	cmd.Flags().StringVar(&cdp, "cdp", "", "Chrome DevTools base URL (default http://localhost:9222)")
	cmd.Flags().IntVar(&logLines, "log-lines", 200, "number of log lines to capture per squawk")
	return cmd
}

func runInit(ctx context.Context, sessionDir, backend, device, tab, cdp string, logLines int) error {
	store, err := session.NewStore(sessionDir)
	if err != nil {
		return err
	}
	b, err := newBackend(backend, cdp)
	if err != nil {
		return err
	}
	if err := b.Check(ctx); err != nil {
		return err
	}

	requested := device
	if backend == "browser" {
		requested = tab
	}
	target, err := selectTarget(ctx, b, requested)
	if err != nil {
		return err
	}

	endpoint := ""
	sessTarget := target.ID
	display := target.ID
	if backend == "browser" {
		endpoint = cdp
		sessTarget = target.URL
		display = target.URL
	}
	sess, err := store.CreateSession(backend, sessTarget, endpoint, logLines)
	if err != nil {
		return err
	}
	fmt.Printf("squawk: session %s started on %s\n", sess.ID, display)
	fmt.Println("squawk: run `squawk watch` or `squawk capture --note \"...\"` to capture bugs")
	fmt.Println("squawk: run `squawk report` to generate a report")
	return nil
}
