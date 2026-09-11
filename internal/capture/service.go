// Package capture coordinates a single squawk capture: it resolves the active
// session and backend, captures the screenshot and log excerpt, persists the
// artifacts, and records the squawk in the session.
package capture

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DRMCGT/Squawk/internal/adb"
	"github.com/DRMCGT/Squawk/internal/browser"
	"github.com/DRMCGT/Squawk/internal/model"
	"github.com/DRMCGT/Squawk/internal/session"
)

// Backend captures screenshots and logs from a target. Implementations wrap
// adb (android) or Chrome DevTools (browser).
type Backend interface {
	// Check verifies the backend is reachable.
	Check(ctx context.Context) error
	// Targets lists selectable targets (devices or open tabs).
	Targets(ctx context.Context) ([]Target, error)
	// Screenshot captures the given target as PNG bytes.
	Screenshot(ctx context.Context, target string) ([]byte, error)
	// Logs returns up to lines recent log entries for the target.
	Logs(ctx context.Context, target string, lines int) ([]byte, error)
}

// Target is a selectable capture target: an adb device or a browser tab.
type Target struct {
	ID    string
	Label string
	URL   string
}

// Service coordinates captures against a session store and a backend.
type Service struct {
	Backend Backend
	Store   *session.Store
}

// Request describes a single capture. Empty Target and LogLines fields fall
// back to the active session's configuration.
type Request struct {
	Note     string
	Target   string
	LogLines int
}

// Capture captures one squawk in the active session and records it.
func (s *Service) Capture(ctx context.Context, req Request) (*model.Squawk, error) {
	sess, err := s.Store.CurrentSession()
	if err != nil {
		return nil, err
	}

	target := req.Target
	if target == "" {
		target = sess.Target
	}
	if target == "" {
		return nil, errors.New("no target configured; run `squawk init` or pass --target")
	}

	logLines := req.LogLines
	if logLines <= 0 {
		logLines = sess.LogLines
	}

	shot, err := s.Backend.Screenshot(ctx, target)
	if err != nil {
		return nil, err
	}
	logOut, err := s.Backend.Logs(ctx, target, logLines)
	if err != nil {
		return nil, err
	}

	id := session.NextSquawkID(sess)
	dir := s.Store.SquawkDir(sess.ID, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create squawk directory: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	if err := os.WriteFile(filepath.Join(dir, "screenshot.png"), shot, 0o644); err != nil {
		cleanup()
		return nil, fmt.Errorf("cannot write screenshot: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "logcat.txt"), logOut, 0o644); err != nil {
		cleanup()
		return nil, fmt.Errorf("cannot write logcat: %w", err)
	}

	sq := &model.Squawk{
		ID:         id,
		CapturedAt: time.Now().UTC(),
		Note:       normalizeNote(req.Note),
		Screenshot: relPath(id, "screenshot.png"),
		Logcat:     relPath(id, "logcat.txt"),
	}

	sess.Squawks = append(sess.Squawks, *sq)
	if err := s.Store.SaveSession(sess); err != nil {
		cleanup()
		return nil, err
	}
	return sq, nil
}

func relPath(id int, name string) string {
	return filepath.ToSlash(filepath.Join("squawks", fmt.Sprintf("%03d", id), name))
}

// normalizeNote trims whitespace and collapses internal runs of whitespace to
// single spaces so notes render cleanly as report titles.
func normalizeNote(note string) string {
	return strings.Join(strings.Fields(note), " ")
}

// ADBBackend adapts an adb client to the Backend interface.
func ADBBackend(c *adb.Client) Backend { return &adbBackend{client: c} }

type adbBackend struct{ client *adb.Client }

func (a *adbBackend) Check(ctx context.Context) error { return a.client.Check(ctx) }

func (a *adbBackend) Targets(ctx context.Context) ([]Target, error) {
	devices, err := a.client.Devices(ctx)
	if err != nil {
		return nil, err
	}
	targets := make([]Target, 0, len(devices))
	for _, d := range devices {
		targets = append(targets, Target{ID: d.Serial, Label: d.State})
	}
	return targets, nil
}

func (a *adbBackend) Screenshot(ctx context.Context, target string) ([]byte, error) {
	return a.client.Screenshot(ctx, target)
}

func (a *adbBackend) Logs(ctx context.Context, target string, lines int) ([]byte, error) {
	return a.client.LogcatTail(ctx, target, lines)
}

// BrowserBackend adapts a browser client to the Backend interface.
func BrowserBackend(c *browser.Client) Backend { return &browserBackend{client: c} }

type browserBackend struct{ client *browser.Client }

func (b *browserBackend) Check(ctx context.Context) error { return b.client.Check(ctx) }

func (b *browserBackend) Targets(ctx context.Context) ([]Target, error) {
	tabs, err := b.client.Tabs(ctx)
	if err != nil {
		return nil, err
	}
	targets := make([]Target, 0, len(tabs))
	for _, t := range tabs {
		targets = append(targets, Target{ID: t.ID, Label: t.Title, URL: t.URL})
	}
	return targets, nil
}

func (b *browserBackend) Screenshot(ctx context.Context, target string) ([]byte, error) {
	return b.client.Screenshot(ctx, target)
}

func (b *browserBackend) Logs(ctx context.Context, target string, lines int) ([]byte, error) {
	return b.client.Logs(ctx, target, lines)
}
