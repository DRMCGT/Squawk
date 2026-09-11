package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squawk-dev/squawk/internal/adb"
	"github.com/squawk-dev/squawk/internal/capture"
	"github.com/squawk-dev/squawk/internal/session"
	"golang.org/x/term"
)

func newCaptureCmd() *cobra.Command {
	var sessionDir, device, note string
	var logLines int

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a squawk: screenshot + recent logcat + note",
		Long: "Capture a squawk in the active session.\n" +
			"Pass --note, or pipe a note via stdin, to attach context without an interactive prompt.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if note == "" && !term.IsTerminal(int(os.Stdin.Fd())) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("cannot read note from stdin: %w", err)
				}
				note = strings.TrimSpace(string(b))
			}
			return runCapture(cmd.Context(), sessionDir, device, note, logLines)
		},
	}
	cmd.Flags().StringVar(&sessionDir, "session-dir", "", "root directory for Squawk data (default ./.squawk)")
	cmd.Flags().StringVar(&device, "device", "", "override the session's configured ADB device")
	cmd.Flags().IntVar(&logLines, "log-lines", 0, "number of logcat lines to capture (default: session setting)")
	cmd.Flags().StringVar(&note, "note", "", "short bug note")
	return cmd
}

func runCapture(ctx context.Context, sessionDir, device, note string, logLines int) error {
	store, err := session.NewStore(sessionDir)
	if err != nil {
		return err
	}
	svc := capture.Service{
		Adb:   adb.NewClient(adb.NewExecRunner()),
		Store: store,
	}
	sq, err := svc.Capture(ctx, capture.Request{Note: note, Device: device, LogLines: logLines})
	if err != nil {
		return err
	}
	fmt.Printf("squawk %03d captured at %s\n", sq.ID, sq.CapturedAt.Format(time.RFC3339))
	return nil
}
