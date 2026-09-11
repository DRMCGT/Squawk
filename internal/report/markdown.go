// Package report renders a session into Markdown or JSON. It reads squawk
// artifacts (logcat excerpts) from the session directory and never invokes
// adb.
package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DRMCGT/Squawk/internal/model"
)

// NoteTitle builds the heading for a squawk, falling back to
// "Squawk 00N (no note)" when the note is empty.
func NoteTitle(sq model.Squawk) string {
	if strings.TrimSpace(sq.Note) == "" {
		return fmt.Sprintf("Squawk %03d (no note)", sq.ID)
	}
	return fmt.Sprintf("Squawk %03d: %s", sq.ID, sq.Note)
}

// Markdown renders a session as a Markdown report. Screenshots are referenced
// with paths relative to the session directory so the report is portable.
func Markdown(sessionDir string, sess *model.Session) ([]byte, error) {
	var b strings.Builder
	b.WriteString("# Squawk Session\n\n")
	fmt.Fprintf(&b, "- Session: `%s`\n", sess.ID)
	fmt.Fprintf(&b, "- Started: `%s`\n", sess.StartedAt.UTC().Format(time.RFC3339))
	if sess.Device != "" {
		fmt.Fprintf(&b, "- Device: `%s`\n", sess.Device)
	}

	if len(sess.Squawks) == 0 {
		b.WriteString("\n_No squawks captured._\n")
		return []byte(b.String()), nil
	}

	for _, sq := range sess.Squawks {
		b.WriteString("\n## " + NoteTitle(sq) + "\n\n")
		fmt.Fprintf(&b, "- Captured: `%s`\n\n", sq.CapturedAt.UTC().Format(time.RFC3339))
		if sq.Screenshot != "" {
			fmt.Fprintf(&b, "![Screenshot](%s)\n\n", sq.Screenshot)
		}
		if sq.Note != "" {
			b.WriteString("### Tester Note\n\n")
			b.WriteString(sq.Note + "\n\n")
		}
		b.WriteString("### Recent Logcat\n\n")
		block, err := logcatBlock(sessionDir, sq.Logcat)
		if err != nil {
			b.WriteString("_log file missing_\n\n")
		} else {
			b.WriteString(block)
		}
	}

	return []byte(b.String()), nil
}

// logcatBlock reads a squawk's logcat excerpt and wraps it in a four-backtick
// fence so triple-backtick sequences inside the content cannot close the
// block early. Backticks in the content are never altered. Line endings are
// normalized to \n and the content is guaranteed to end in a newline.
func logcatBlock(sessionDir, relPath string) (string, error) {
	if relPath == "" {
		return "", os.ErrNotExist
	}
	data, err := os.ReadFile(filepath.Join(sessionDir, filepath.FromSlash(relPath)))
	if err != nil {
		return "", err
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return "````text\n" + content + "````\n", nil
}
