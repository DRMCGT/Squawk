package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/DRMCGT/Squawk/internal/model"
	"github.com/DRMCGT/Squawk/internal/session"
	"github.com/spf13/cobra"
)

func newDevicesCmd() *cobra.Command {
	var backend, cdp string
	cmd := &cobra.Command{
		Use:   "devices",
		Short: "List connected targets (Android devices or browser tabs)",
		Long: "Lists Android devices by default, or open browser tabs for a browser session\n" +
			"(or when --backend browser is passed).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDevices(cmd.Context(), backend, cdp)
		},
	}
	cmd.Flags().StringVar(&backend, "backend", "", "list targets for this backend (android or browser); defaults to the active session's backend")
	cmd.Flags().StringVar(&cdp, "cdp", "", "Chrome DevTools base URL (default http://localhost:9222)")
	return cmd
}

func runDevices(ctx context.Context, backend, cdp string) error {
	if backend == "" {
		if store, err := session.NewStore(""); err == nil {
			if sess, err := store.CurrentSession(); err == nil {
				backend = sess.Backend
				if cdp == "" {
					cdp = sess.Endpoint
				}
			}
		}
	}
	if backend == "" {
		backend = model.BackendAndroid
	}

	b, err := newBackend(backend, cdp)
	if err != nil {
		return err
	}
	targets, err := b.Targets(ctx)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		if backend == model.BackendBrowser {
			fmt.Println("no browser tabs are open")
		} else {
			fmt.Println("no Android devices are connected")
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if backend == model.BackendBrowser {
		fmt.Fprintln(w, "TAB\tURL")
		for _, t := range targets {
			fmt.Fprintf(w, "%s\t%s\n", t.Label, t.URL)
		}
	} else {
		fmt.Fprintln(w, "DEVICE\tSTATE")
		for _, t := range targets {
			fmt.Fprintf(w, "%s\t%s\n", t.ID, t.Label)
		}
	}
	return w.Flush()
}
