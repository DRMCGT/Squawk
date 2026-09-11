package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squawk-dev/squawk/internal/adb"
	"github.com/squawk-dev/squawk/internal/capture"
	"github.com/squawk-dev/squawk/internal/session"
)

func newWatchCmd() *cobra.Command {
	var sessionDir, device string
	var logLines int

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Interactive capture loop: press Enter to capture, type q to quit",
		Long: "Run an interactive capture loop against the active session.\n" +
			"Press Enter to capture (you will be prompted for an optional note), or type a note and press Enter to capture it directly.\n" +
			"Type q and press Enter (or press Ctrl-C / close stdin) to exit.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWatch(cmd.Context(), sessionDir, device, logLines)
		},
	}
	cmd.Flags().StringVar(&sessionDir, "session-dir", "", "root directory for Squawk data (default ./.squawk)")
	cmd.Flags().StringVar(&device, "device", "", "override the session's configured ADB device")
	cmd.Flags().IntVar(&logLines, "log-lines", 0, "number of logcat lines to capture (default: session setting)")
	return cmd
}

func runWatch(ctx context.Context, sessionDir, device string, logLines int) error {
	store, err := session.NewStore(sessionDir)
	if err != nil {
		return err
	}
	if _, err := store.CurrentSession(); err != nil {
		return err
	}
	svc := capture.Service{
		Adb:   adb.NewClient(adb.NewExecRunner()),
		Store: store,
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("squawk watch: press Enter to capture a bug, type a note then Enter to capture with it, type q to quit, Ctrl-C to exit")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stdout)
			return nil
		}
		input := strings.TrimSpace(line)
		switch input {
		case "q", "quit", "exit":
			return nil
		}

		note := input
		if note == "" {
			fmt.Print("note (optional): ")
			nb, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintln(os.Stdout)
				return nil
			}
			note = strings.TrimSpace(nb)
		}

		sq, err := svc.Capture(ctx, capture.Request{Note: note, Device: device, LogLines: logLines})
		if err != nil {
			fmt.Fprintf(os.Stderr, "squawk: capture failed: %v\n", err)
			continue
		}
		fmt.Printf("squawk %03d captured\n", sq.ID)
	}
}
