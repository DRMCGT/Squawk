package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	squawkBin   string
	fakeAdbPath string
)

func TestMain(m *testing.M) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		os.Exit(1)
	}
	binDir, err := os.MkdirTemp("", "squawk-bin-")
	if err != nil {
		os.Exit(1)
	}
	squawkBin = filepath.Join(binDir, "squawk")
	cmd := exec.Command("go", "build", "-o", squawkBin, "./cmd/squawk")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Stderr.Write(out)
		os.Exit(1)
	}
	fakeAdbPath = filepath.Join(repoRoot, "testdata", "fake-adb", "adb")
	code := m.Run()
	_ = os.RemoveAll(binDir)
	os.Exit(code)
}

func runSquawk(t *testing.T, sessionDir string, stdin string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(squawkBin, args...)
	cmd.Env = append(os.Environ(), "SQUAWK_ADB="+fakeAdbPath)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errBuf strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return out.String(), errBuf.String(), err
}

func TestEndToEndWorkflow(t *testing.T) {
	sessionDir := t.TempDir()

	out, _, err := runSquawk(t, sessionDir, "", "devices")
	if err != nil {
		t.Fatalf("devices: %v", err)
	}
	if !strings.Contains(out, "emulator-5554") || !strings.Contains(out, "device") {
		t.Fatalf("devices output unexpected:\n%s", out)
	}

	out, errOut, err := runSquawk(t, sessionDir, "", "init", "--session-dir", sessionDir)
	if err != nil {
		t.Fatalf("init: %v (stderr: %s)", err, errOut)
	}
	if !strings.Contains(out, "started on emulator-5554") {
		t.Fatalf("init output unexpected:\n%s", out)
	}

	out, _, err = runSquawk(t, sessionDir, "", "capture", "--session-dir", sessionDir, "--note", "save broken")
	if err != nil {
		t.Fatalf("capture 1: %v", err)
	}
	if !strings.Contains(out, "squawk 001 captured") {
		t.Fatalf("capture 1 output unexpected:\n%s", out)
	}

	out, _, err = runSquawk(t, sessionDir, "  ", "capture", "--session-dir", sessionDir)
	if err != nil {
		t.Fatalf("capture 2: %v", err)
	}
	if !strings.Contains(out, "squawk 002 captured") {
		t.Fatalf("capture 2 output unexpected:\n%s", out)
	}

	out, errOut, err = runSquawk(t, sessionDir, "", "report", "--session-dir", sessionDir)
	if err != nil {
		t.Fatalf("report: %v (stderr: %s)", err, errOut)
	}
	if !strings.Contains(out, "report.md") {
		t.Fatalf("report output unexpected:\n%s", out)
	}

	reportPath := filepath.Join(sessionDir, "sessions")
	entries, err := os.ReadDir(reportPath)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one session dir, got %v (err %v)", entries, err)
	}
	sessDir := filepath.Join(reportPath, entries[0].Name())
	md, err := os.ReadFile(filepath.Join(sessDir, "report.md"))
	if err != nil {
		t.Fatalf("read report.md: %v", err)
	}
	body := string(md)
	for _, want := range []string{
		"## Squawk 001: save broken",
		"## Squawk 002 (no note)",
		"![Screenshot](squawks/001/screenshot.png)",
		"````text",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("report.md missing %q:\n%s", want, body)
		}
	}

	out, _, err = runSquawk(t, sessionDir, "", "report", "--session-dir", sessionDir, "--format", "json")
	if err != nil {
		t.Fatalf("report json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sessDir, "report.json")); err != nil {
		t.Fatalf("report.json missing: %v", err)
	}

	out, _, err = runSquawk(t, sessionDir, "", "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(out, "squawk version") {
		t.Fatalf("version output unexpected:\n%s", out)
	}
}

func TestNoActiveSessionError(t *testing.T) {
	sessionDir := t.TempDir()
	_, errOut, err := runSquawk(t, sessionDir, "", "report", "--session-dir", sessionDir)
	if err == nil {
		t.Fatal("expected error for report without active session")
	}
	if !strings.Contains(errOut, "no active Squawk session found") {
		t.Fatalf("stderr unexpected:\n%s", errOut)
	}
}

func TestCaptureStdinNote(t *testing.T) {
	sessionDir := t.TempDir()
	if _, _, err := runSquawk(t, sessionDir, "", "init", "--session-dir", sessionDir); err != nil {
		t.Fatalf("init: %v", err)
	}
	out, _, err := runSquawk(t, sessionDir, "save button unresponsive\n", "capture", "--session-dir", sessionDir)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if !strings.Contains(out, "squawk 001 captured") {
		t.Fatalf("capture output unexpected:\n%s", out)
	}
}
