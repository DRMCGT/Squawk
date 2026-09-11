package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squawk-dev/squawk/internal/model"
	"github.com/squawk-dev/squawk/internal/report"
	"github.com/squawk-dev/squawk/internal/session"
)

func newReportCmd() *cobra.Command {
	var sessionDir, sessionID, format, output string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a Markdown or JSON report for a session",
		Long: "Generate a report for the active session (or a specific one with --session).\n" +
			"The report embeds screenshots by relative path and includes the recent logcat excerpts and tester notes.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := session.NewStore(sessionDir)
			if err != nil {
				return err
			}

			var sess *model.Session
			if sessionID != "" {
				sess, err = store.LoadSession(sessionID)
			} else {
				sess, err = store.CurrentSession()
			}
			if err != nil {
				return err
			}

			format = strings.ToLower(strings.TrimSpace(format))
			if format == "" {
				format = "markdown"
			}
			var data []byte
			switch format {
			case "markdown":
				data, err = report.Markdown(store.SessionDir(sess.ID), sess)
			case "json":
				data, err = report.JSON(sess)
			default:
				return fmt.Errorf("unsupported format %q; use markdown or json", format)
			}
			if err != nil {
				return err
			}

			if output == "" {
				ext := ".md"
				if format == "json" {
					ext = ".json"
				}
				output = filepath.Join(store.SessionDir(sess.ID), "report"+ext)
			}
			if err := os.WriteFile(output, data, 0o644); err != nil {
				return fmt.Errorf("cannot write report: %w", err)
			}
			fmt.Printf("squawk: report written to %s\n", output)
			return nil
		},
	}
	cmd.Flags().StringVar(&sessionDir, "session-dir", "", "root directory for Squawk data (default ./.squawk)")
	cmd.Flags().StringVar(&sessionID, "session", "", "session ID to report on (default: active session)")
	cmd.Flags().StringVar(&format, "format", "markdown", "report format: markdown or json")
	cmd.Flags().StringVar(&output, "output", "", "output path (default: <session dir>/report.md or report.json)")
	return cmd
}
