package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/squawk-dev/squawk/internal/model"
)

func writeLogcat(t *testing.T, dir, rel string, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sampleSession() *model.Session {
	now := time.Date(2026, 9, 10, 14, 30, 12, 0, time.UTC)
	return &model.Session{
		ID:        "20260910-143012",
		StartedAt: now,
		Device:    "emulator-5554",
		LogLines:  200,
		Squawks: []model.Squawk{
			{ID: 1, CapturedAt: now.Add(time.Minute), Note: "save button unresponsive", Screenshot: "squawks/001/screenshot.png", Logcat: "squawks/001/logcat.txt"},
			{ID: 2, CapturedAt: now.Add(2 * time.Minute), Note: "", Screenshot: "squawks/002/screenshot.png", Logcat: "squawks/002/logcat.txt"},
		},
	}
}

func TestNoteTitle(t *testing.T) {
	cases := []struct {
		note string
		id   int
		want string
	}{
		{"save button unresponsive", 1, "Squawk 001: save button unresponsive"},
		{"", 3, "Squawk 003 (no note)"},
		{"   ", 3, "Squawk 003 (no note)"},
	}
	for _, c := range cases {
		if got := NoteTitle(model.Squawk{ID: c.id, Note: c.note}); got != c.want {
			t.Fatalf("NoteTitle(%q, %d) = %q, want %q", c.note, c.id, got, c.want)
		}
	}
}

func TestMarkdownContent(t *testing.T) {
	dir := t.TempDir()
	sess := sampleSession()
	logContent := "E/MyApp: boom\nline with ``` triple backticks\n"
	writeLogcat(t, dir, "squawks/001/logcat.txt", logContent)
	writeLogcat(t, dir, "squawks/002/logcat.txt", "E/MyApp: crash\n")

	data, err := Markdown(dir, sess)
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	md := string(data)

	for _, want := range []string{
		"# Squawk Session",
		"- Session: `20260910-143012`",
		"- Device: `emulator-5554`",
		"## Squawk 001: save button unresponsive",
		"![Screenshot](squawks/001/screenshot.png)",
		"## Squawk 002 (no note)",
		"### Recent Logcat",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q:\n%s", want, md)
		}
	}

	// The logcat excerpt containing triple backticks must be preserved inside a
	// four-backtick fence and must not close the block early.
	if !strings.Contains(md, "````text\n"+logContent) {
		t.Fatalf("expected four-backtick fence around raw logcat content:\n%s", md)
	}
	if strings.Contains(md, "\n```\nE/MyApp: boom") {
		t.Fatalf("triple backticks closed the fence early:\n%s", md)
	}
}

func TestMarkdownNoSquawks(t *testing.T) {
	dir := t.TempDir()
	sess := &model.Session{ID: "x", StartedAt: time.Now().UTC(), Squawks: nil}
	data, err := Markdown(dir, sess)
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	if !strings.Contains(string(data), "No squawks captured") {
		t.Fatalf("expected empty-state notice:\n%s", data)
	}
}

func TestMarkdownMissingLogcat(t *testing.T) {
	dir := t.TempDir()
	sess := &model.Session{
		ID: "x", StartedAt: time.Now().UTC(),
		Squawks: []model.Squawk{{ID: 1, CapturedAt: time.Now().UTC(), Logcat: "squawks/001/logcat.txt"}},
	}
	data, err := Markdown(dir, sess)
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	if !strings.Contains(string(data), "_log file missing_") {
		t.Fatalf("expected missing-log notice:\n%s", data)
	}
}

func TestJSONOutput(t *testing.T) {
	sess := sampleSession()
	data, err := JSON(sess)
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	for _, want := range []string{`"id": 2`, `"note": ""`, `"started_at": "2026-09-10T14:30:12Z"`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("json missing %q:\n%s", want, data)
		}
	}
}
