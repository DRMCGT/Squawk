// Package capture coordinates a single squawk capture: it resolves the active
// session and device, captures the screenshot and logcat excerpt, persists the
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
	"github.com/DRMCGT/Squawk/internal/model"
	"github.com/DRMCGT/Squawk/internal/session"
)

// Service coordinates captures against a session store and adb client.
type Service struct {
	Adb   *adb.Client
	Store *session.Store
}

// Request describes a single capture. Empty Device and LogLines fields fall
// back to the active session's configuration.
type Request struct {
	Note     string
	Device   string
	LogLines int
}

// Capture captures one squawk in the active session and records it.
func (s *Service) Capture(ctx context.Context, req Request) (*model.Squawk, error) {
	sess, err := s.Store.CurrentSession()
	if err != nil {
		return nil, err
	}

	device := req.Device
	if device == "" {
		device = sess.Device
	}
	if device == "" {
		return nil, errors.New("no device configured; run `squawk init --device <serial>` or pass --device")
	}

	logLines := req.LogLines
	if logLines <= 0 {
		logLines = sess.LogLines
	}

	shot, err := s.Adb.Screenshot(ctx, device)
	if err != nil {
		return nil, err
	}
	logOut, err := s.Adb.LogcatTail(ctx, device, logLines)
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
