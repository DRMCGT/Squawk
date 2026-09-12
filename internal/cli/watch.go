package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DRMCGT/Squawk/internal/capture"
	"github.com/DRMCGT/Squawk/internal/session"
	"github.com/spf13/cobra"
)

func newWatchCmd() *cobra.Command {
	var sessionDir, target string
	var logLines int

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Interactive capture loop: press Enter to capture, type q to quit",
		Long: "Run an interactive capture loop against the active session.\n" +
			"Press Enter to capture (you will be prompted for an optional note), or type a note and press Enter to capture it directly.\n" +
			"Slash commands: /capture, /note <text>, /report, /help, /quit (or /q).\n" +
			"Type q and press Enter (or press Ctrl-C / close stdin) to exit.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWatch(cmd.Context(), sessionDir, target, logLines)
		},
	}
	cmd.Flags().StringVar(&sessionDir, "session-dir", "", "root directory for Squawk data (default ./.squawk)")
	cmd.Flags().StringVar(&target, "target", "", "override the session's configured target (device serial or tab URL)")
	cmd.Flags().IntVar(&logLines, "log-lines", 0, "number of log lines to capture (default: session setting)")
	return cmd
}

func runWatch(ctx context.Context, sessionDir, target string, logLines int) error {
	store, err := session.NewStore(sessionDir)
	if err != nil {
		return err
	}
	sess, err := store.CurrentSession()
	if err != nil {
		return err
	}
	b, err := backendForSession(sess)
	if err != nil {
		return err
	}
	svc := capture.Service{
		Backend: b,
		Store:   store,
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("squawk watch: press Enter or /capture to capture a bug, type a note then Enter to capture with it, /note <text> for a one-line note, /report to refresh the report, /help for commands, q or /quit to exit")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stdout)
			return nil
		}
		input := strings.TrimSpace(line)
		switch input {
		case "q", "quit", "exit", "/quit", "/q":
			return nil
		case "/help":
			fmt.Println(watchHelp)
			continue
		case "/report":
			if err := runReport(ctx, sessionDir, "", "markdown", ""); err != nil {
				fmt.Fprintf(os.Stderr, "squawk: report failed: %v\n", err)
			}
			continue
		case "/capture":
			note := promptNote(reader)
			captureOne(ctx, svc, note, target, logLines)
			continue
		case "/note":
			fmt.Fprintln(os.Stderr, "squawk: /note requires text; use `/note <text>`")
			continue
		}

		if strings.HasPrefix(input, "/note ") {
			note := strings.TrimSpace(strings.TrimPrefix(input, "/note "))
			captureOne(ctx, svc, note, target, logLines)
			continue
		}
		if strings.HasPrefix(input, "/") {
			fmt.Fprintf(os.Stderr, "squawk: unknown command %q; type /help for available commands\n", input)
			continue
		}

		if input == "" {
			captureOne(ctx, svc, promptNote(reader), target, logLines)
			continue
		}
		captureOne(ctx, svc, input, target, logLines)
	}
}

// watchHelp lists the slash commands available inside `squawk watch`.
const watchHelp = `commands:
  (Enter)          capture a bug (you will be prompted for an optional note)
  /capture         capture a bug (same as Enter)
  /note <text>     capture with a note directly, no separate prompt
  /report          regenerate the session report without leaving watch
  /help            show this help
  /quit, /q        quit watch (q, quit, exit also work)`

// promptNote asks for an optional note and returns it trimmed.
func promptNote(reader *bufio.Reader) string {
	fmt.Print("note (optional): ")
	nb, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stdout)
		return ""
	}
	return strings.TrimSpace(nb)
}

// captureOne runs a single capture and reports the result (or error) inline.
func captureOne(ctx context.Context, svc capture.Service, note, target string, logLines int) {
	sq, err := svc.Capture(ctx, capture.Request{Note: note, Target: target, LogLines: logLines})
	if err != nil {
		fmt.Fprintf(os.Stderr, "squawk: capture failed: %v\n", err)
		return
	}
	fmt.Printf("squawk %03d captured\n", sq.ID)
}
