package capture

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DRMCGT/Squawk/internal/adb"
	"github.com/DRMCGT/Squawk/internal/session"
)

var pngBytes = append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, []byte("image")...)

type stubRunner struct {
	shot  []byte
	logs  []byte
	order []string
	err   error
}

func (s *stubRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	s.order = append(s.order, strings.Join(args, " "))
	if s.err != nil {
		return nil, s.err
	}
	for _, a := range args {
		switch a {
		case "screencap":
			return s.shot, nil
		case "logcat":
			return s.logs, nil
		}
	}
	return nil, os.ErrNotExist
}

func newTestService(t *testing.T) (*Service, *session.Store) {
	t.Helper()
	root := t.TempDir()
	store, err := session.NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, err := store.CreateSession("android", "emulator-5554", "", 200); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return &Service{Backend: ADBBackend(adb.NewClient(&stubRunner{shot: pngBytes, logs: []byte("log line\n")})), Store: store}, store
}

func TestCaptureSuccess(t *testing.T) {
	svc, store := newTestService(t)
	sq, err := svc.Capture(context.Background(), Request{Note: "  save button unresponsive  "})
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if sq.ID != 1 {
		t.Fatalf("expected ID 1, got %d", sq.ID)
	}
	if sq.Note != "save button unresponsive" {
		t.Fatalf("expected trimmed note, got %q", sq.Note)
	}
	if sq.Screenshot != "squawks/001/screenshot.png" || sq.Logcat != "squawks/001/logcat.txt" {
		t.Fatalf("unexpected paths: %q %q", sq.Screenshot, sq.Logcat)
	}

	sess, err := store.CurrentSession()
	if err != nil {
		t.Fatalf("CurrentSession: %v", err)
	}
	dir := store.SessionDir(sess.ID)
	if _, err := os.Stat(filepath.Join(dir, "squawks", "001", "screenshot.png")); err != nil {
		t.Fatalf("screenshot artifact missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "squawks", "001", "logcat.txt")); err != nil {
		t.Fatalf("logcat artifact missing: %v", err)
	}
	if len(sess.Squawks) != 1 {
		t.Fatalf("expected 1 squawk in session, got %d", len(sess.Squawks))
	}

	sq2, err := svc.Capture(context.Background(), Request{Note: ""})
	if err != nil {
		t.Fatalf("Capture 2: %v", err)
	}
	if sq2.ID != 2 {
		t.Fatalf("expected ID 2, got %d", sq2.ID)
	}
	if sq2.Note != "" {
		t.Fatalf("expected empty note preserved, got %q", sq2.Note)
	}
	sess, _ = store.CurrentSession()
	if len(sess.Squawks) != 2 {
		t.Fatalf("expected 2 squawks, got %d", len(sess.Squawks))
	}
}

func TestCaptureNoActiveSession(t *testing.T) {
	root := t.TempDir()
	store, err := session.NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	svc := &Service{Backend: ADBBackend(adb.NewClient(&stubRunner{shot: pngBytes, logs: []byte("x")})), Store: store}
	if _, err := svc.Capture(context.Background(), Request{}); err == nil {
		t.Fatal("expected error for missing active session")
	}
}

func TestCaptureScreenshotFailureCleansUp(t *testing.T) {
	svc, _ := newTestService(t)
	svc.Backend = ADBBackend(adb.NewClient(&stubRunner{err: os.ErrNotExist}))
	if _, err := svc.Capture(context.Background(), Request{}); err == nil {
		t.Fatal("expected capture failure")
	}
	sess, _ := svc.Store.CurrentSession()
	if len(sess.Squawks) != 0 {
		t.Fatalf("expected no squawks recorded after failure, got %d", len(sess.Squawks))
	}
}
