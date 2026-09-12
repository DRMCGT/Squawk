package cli

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/DRMCGT/Squawk/internal/capture"
	"github.com/DRMCGT/Squawk/internal/model"
	"github.com/DRMCGT/Squawk/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

type stubBackend struct{}

func (stubBackend) Check(context.Context) error { return nil }

func (stubBackend) Targets(context.Context) ([]capture.Target, error) {
	return []capture.Target{{ID: "stub-1", Label: "device"}}, nil
}

func (stubBackend) Screenshot(context.Context, string) ([]byte, error) {
	return []byte("\x89PNG\r\n\x1a\nfake-png-bytes"), nil
}

func (stubBackend) Logs(context.Context, string, int) ([]byte, error) {
	return []byte("[00:00:00] console.log: stub\n"), nil
}

func newTUIForTest(t *testing.T) (watchModel, *session.Store, *model.Session) {
	t.Helper()
	store, err := session.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	sess, err := store.CreateSession(model.BackendAndroid, "stub-1", "", 200)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	svc := capture.Service{Backend: stubBackend{}, Store: store}
	m := newWatchModel(context.Background(), store, sess, svc, "stub-1", 0)
	m.width, m.height = 80, 24
	return m, store, sess
}

// sendLine drives one submit through the real Update path and executes the
// returned command so capture/report results are applied.
func sendLine(t *testing.T, m *watchModel, line string) {
	t.Helper()
	m.input = line
	updated, cmd := m.submit()
	nm, ok := updated.(watchModel)
	if !ok {
		t.Fatalf("submit did not return watchModel")
	}
	*m = nm
	if cmd == nil {
		return
	}
	if msg := cmd(); msg != nil {
		if _, isQuit := msg.(tea.QuitMsg); isQuit {
			return
		}
		updated, _ = m.Update(msg)
		nm, ok = updated.(watchModel)
		if !ok {
			t.Fatalf("async message did not return watchModel")
		}
		*m = nm
	}
}

func TestTUIInitialView(t *testing.T) {
	m, _, _ := newTUIForTest(t)
	view := m.View()
	for _, want := range []string{
		"______",
		"\\___/\\__,_/ |___/",
		"Session",
		"Backend android",
		"stub-1",
		"Squawks 0",
		"/capture",
		"/note <text>",
		"/report",
		"/help",
		"Type a note, or /help for commands",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("initial view missing %q:\n%s", want, view)
		}
	}
	if _, err := os.Stat(m.store.SessionDir(m.sess.ID)); err != nil {
		t.Fatalf("session dir: %v", err)
	}
}

func TestTUISlashNoteCaptures(t *testing.T) {
	m, store, _ := newTUIForTest(t)
	sendLine(t, &m, "/note save button broken")
	if m.status != "Squawk 001 captured" {
		t.Fatalf("status after capture = %q", m.status)
	}
	fresh, err := store.CurrentSession()
	if err != nil || len(fresh.Squawks) != 1 || fresh.Squawks[0].Note != "save button broken" {
		t.Fatalf("session after capture: %+v (err %v)", fresh, err)
	}
}

func TestTUIBareTextAndEnter(t *testing.T) {
	m, store, _ := newTUIForTest(t)
	sendLine(t, &m, "bare note")
	if m.status != "Squawk 001 captured" {
		t.Fatalf("bare text capture: %q", m.status)
	}
	sendLine(t, &m, "")
	if !m.awaitNote {
		t.Fatal("empty Enter should enter optional-note mode")
	}
	sendLine(t, &m, "second note")
	if m.status != "Squawk 002 captured" {
		t.Fatalf("awaited capture: %q", m.status)
	}
	fresh, _ := store.CurrentSession()
	if len(fresh.Squawks) != 2 {
		t.Fatalf("expected 2 squawks, got %d", len(fresh.Squawks))
	}
}

func TestTUICaptureCommand(t *testing.T) {
	m, _, _ := newTUIForTest(t)
	sendLine(t, &m, "/capture")
	if !m.awaitNote {
		t.Fatal("/capture should enter optional-note mode")
	}
	sendLine(t, &m, "from capture cmd")
	if m.status != "Squawk 001 captured" {
		t.Fatalf("/capture flow: %q", m.status)
	}
}

func TestTUIReportCommand(t *testing.T) {
	m, store, sess := newTUIForTest(t)
	sendLine(t, &m, "/note for the report")
	sendLine(t, &m, "/report")
	if !strings.HasPrefix(m.status, "Report written to") {
		t.Fatalf("report status = %q (err %q)", m.status, m.errLine)
	}
	data, err := os.ReadFile(store.SessionDir(sess.ID) + "/report.md")
	if err != nil {
		t.Fatalf("report.md: %v", err)
	}
	if !strings.Contains(string(data), "for the report") {
		t.Fatalf("report.md missing note:\n%s", data)
	}
}

func TestTUIHelpQuitsAndErrors(t *testing.T) {
	m, _, _ := newTUIForTest(t)

	sendLine(t, &m, "/help")
	if !m.showHelp {
		t.Fatal("/help should show help panel")
	}
	if !strings.Contains(m.View(), "commands:") {
		t.Fatal("view missing help panel")
	}

	sendLine(t, &m, "/bogus")
	if m.showHelp {
		t.Fatal("help should dismiss on next command")
	}
	if !strings.Contains(m.errLine, "unknown command") {
		t.Fatalf("expected unknown-command error, got %q", m.errLine)
	}

	sendLine(t, &m, "/note")
	if !strings.Contains(m.errLine, "requires text") {
		t.Fatalf("expected /note usage error, got %q", m.errLine)
	}

	if m.quitting {
		t.Fatal("should not have quit yet")
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl-c must return a command")
	}
	msg := cmd()
	if msg != (tea.QuitMsg{}) {
		t.Fatalf("ctrl-c msg = %#v", msg)
	}
	if nm, ok := updated.(watchModel); !ok || !nm.quitting {
		t.Fatal("ctrl-c should mark quitting")
	}
}

func TestTUICaptureErrorKeepsRunning(t *testing.T) {
	m, store, _ := newTUIForTest(t)
	failing := capture.Service{Backend: failBackend{}, Store: store}
	m.svc = failing
	sendLine(t, &m, "/note boom")
	if !strings.Contains(m.errLine, "capture failed") {
		t.Fatalf("expected inline capture error, got %q", m.errLine)
	}
	if m.quitting || m.capturing {
		t.Fatal("model should stay usable after a failed capture")
	}
}

type failBackend struct{ stubBackend }

func (failBackend) Screenshot(context.Context, string) ([]byte, error) {
	return nil, errors.New("screenshot boom")
}

func TestTUICapturingStateInView(t *testing.T) {
	m, _, _ := newTUIForTest(t)
	m.input = "/note slow capture"
	updated, cmd := m.submit()
	m = updated.(watchModel)
	if cmd == nil {
		t.Fatal("expected a capture command")
	}
	if !m.capturing {
		t.Fatal("model should be capturing before the command resolves")
	}
	view := m.View()
	if !strings.Contains(view, "Capturing…") {
		t.Fatalf("view should show Capturing state:\n%s", view)
	}
	if strings.Contains(view, "Type a note") {
		t.Fatal("input box should be replaced while capturing")
	}
}

func TestTUIWritingReportStateInView(t *testing.T) {
	m, _, _ := newTUIForTest(t)
	m.input = "/report"
	updated, cmd := m.submit()
	m = updated.(watchModel)
	if cmd == nil {
		t.Fatal("expected a report command")
	}
	if !m.reporting {
		t.Fatal("model should be reporting before the command resolves")
	}
	if !strings.Contains(m.View(), "Writing report…") {
		t.Fatal("view should show Writing report state")
	}
}

func TestTUIInputIgnoredWhileCapturing(t *testing.T) {
	m, _, _ := newTUIForTest(t)
	m.capturing = true
	m.input = "old"
	updated, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("xyz")})
	if nm, ok := updated.(watchModel); !ok || nm.input != "old" {
		t.Fatalf("input changed while capturing: %q", updated.(watchModel).input)
	}
	if cmd != nil {
		t.Fatal("no command expected while capturing")
	}
}

func TestTUISpaceKey(t *testing.T) {
	m, store, _ := newTUIForTest(t)
	for _, msg := range []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/note hello")},
		tea.KeyMsg{Type: tea.KeySpace},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("world")},
		tea.KeyMsg{Type: tea.KeyEnter},
	} {
		updated, cmd := m.Update(msg)
		m = updated.(watchModel)
		if cmd != nil {
			switch res := cmd().(type) {
			case captureDoneMsg, reportDoneMsg:
				updated, _ = m.Update(res)
				m = updated.(watchModel)
			}
		}
	}
	if m.status != "Squawk 001 captured" {
		t.Fatalf("space input flow: status=%q err=%q", m.status, m.errLine)
	}
	fresh, _ := store.CurrentSession()
	if len(fresh.Squawks) != 1 || fresh.Squawks[0].Note != "hello world" {
		t.Fatalf("note with space not persisted: %+v", fresh.Squawks)
	}
}
