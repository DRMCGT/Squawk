package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/DRMCGT/Squawk/internal/adb"
	"github.com/DRMCGT/Squawk/internal/browser"
	"github.com/DRMCGT/Squawk/internal/capture"
	"github.com/DRMCGT/Squawk/internal/model"
)

// newBackend builds a capture.Backend from a backend name and, for browser,
// a Chrome DevTools base URL.
func newBackend(name, cdp string) (capture.Backend, error) {
	switch name {
	case model.BackendAndroid, "":
		return capture.ADBBackend(adb.NewClient(adb.NewExecRunner())), nil
	case model.BackendBrowser:
		return capture.BrowserBackend(browser.New(cdp)), nil
	default:
		return nil, fmt.Errorf("unknown backend %q; use %q or %q", name, model.BackendAndroid, model.BackendBrowser)
	}
}

// backendForSession returns the backend matching a session's stored
// configuration.
func backendForSession(sess *model.Session) (capture.Backend, error) {
	return newBackend(sess.Backend, sess.Endpoint)
}

// selectTarget resolves a capture target for a backend from an optional
// requested selector, mirroring adb device selection. For browser backends a
// selector may be a URL substring.
func selectTarget(ctx context.Context, b capture.Backend, requested string) (capture.Target, error) {
	targets, err := b.Targets(ctx)
	if err != nil {
		return capture.Target{}, err
	}
	if requested != "" {
		for _, t := range targets {
			if t.ID == requested {
				return t, nil
			}
		}
		var matches []capture.Target
		for _, t := range targets {
			if strings.Contains(t.URL, requested) {
				matches = append(matches, t)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 {
			return capture.Target{}, fmt.Errorf("multiple targets match %q; use a more specific selector", requested)
		}
		return capture.Target{}, fmt.Errorf("target %q not found", requested)
	}
	if len(targets) == 0 {
		return capture.Target{}, fmt.Errorf("no targets available; connect an Android device or open a browser tab")
	}
	if len(targets) > 1 {
		return capture.Target{}, fmt.Errorf("multiple targets available; specify one with --device/--tab")
	}
	return targets[0], nil
}
