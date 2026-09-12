package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DRMCGT/Squawk/internal/capture"
	"github.com/DRMCGT/Squawk/internal/model"
	"github.com/DRMCGT/Squawk/internal/session"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// squawkLogo is the block-letter wordmark shown at the top of the watch TUI.
const squawkLogo = "   ______                    __\n" +
	"  /_  __/___  ____ __  __    / /\n" +
	"   / / / _ \\/ __ `/ | / /   / /\n" +
	"  / / /  __/ /_/ /| |/ /   / /__\n" +
	" /_/   \\___/\\__,_/ |___/   /_____/"

var (
	tuiAccent     = lipgloss.Color("208")
	tuiDim        = lipgloss.Color("243")
	tuiFaint      = lipgloss.Color("240")
	tuiSuccess    = lipgloss.Color("42")
	tuiErrorColor = lipgloss.Color("203")

	tuiLogoStyle = lipgloss.NewStyle().
			Foreground(tuiAccent).
			Bold(true)

	tuiStatusStyle = lipgloss.NewStyle().
			Foreground(tuiDim)

	tuiPromptStyle = lipgloss.NewStyle().
			Foreground(tuiAccent).
			Bold(true)

	tuiPlaceholderStyle = lipgloss.NewStyle().
				Foreground(tuiFaint).
				Italic(true)

	tuiCursorStyle = lipgloss.NewStyle().
			Reverse(true)

	tuiBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(tuiDim).
			Padding(0, 1)

	tuiFooterStyle = lipgloss.NewStyle().
			Foreground(tuiFaint)

	tuiSuccessStyle = lipgloss.NewStyle().
			Foreground(tuiSuccess)

	tuiErrorStyle = lipgloss.NewStyle().
			Foreground(tuiErrorColor)

	tuiBusyStyle = lipgloss.NewStyle().
			Foreground(tuiAccent).
			Italic(true)

	tuiHelpStyle = lipgloss.NewStyle().
			Foreground(tuiDim).
			PaddingLeft(2)
)

// captureDoneMsg is the async result of a capture running as a tea.Cmd.
type captureDoneMsg struct {
	id  int
	err error
}

// reportDoneMsg is the async result of `/report` running as a tea.Cmd.
type reportDoneMsg struct {
	path string
	err  error
}

// fadeStatusMsg clears the transient confirmation line a few seconds after a
// successful capture or report.
type fadeStatusMsg struct{}

// statusFadeDelay is how long a confirmation stays visible above the input.
const statusFadeDelay = 3 * time.Second

// watchModel is the full-screen TUI for `squawk watch`. It renders over the
// existing capture/report services and keeps the same command set as the
// plain-text loop; it is only used when watch runs on a real terminal.
type watchModel struct {
	ctx      context.Context
	store    *session.Store
	svc      capture.Service
	sess     *model.Session
	target   string
	logLines int

	width, height int

	input     string
	awaitNote bool
	capturing bool
	reporting bool
	status    string
	errLine   string
	showHelp  bool
	quitting  bool
}

func newWatchModel(ctx context.Context, store *session.Store, sess *model.Session, svc capture.Service, target string, logLines int) watchModel {
	return watchModel{
		ctx:      ctx,
		store:    store,
		svc:      svc,
		sess:     sess,
		target:   target,
		logLines: logLines,
	}
}

func (m watchModel) Init() tea.Cmd { return nil }

func (m watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case captureDoneMsg:
		m.capturing = false
		if msg.err != nil {
			m.status = ""
			m.errLine = fmt.Sprintf("capture failed: %v", msg.err)
			return m, nil
		}
		m.errLine = ""
		m.status = fmt.Sprintf("Squawk %03d captured", msg.id)
		if fresh, err := m.store.CurrentSession(); err == nil {
			m.sess = fresh
		}
		return m, fadeStatusCmd()
	case reportDoneMsg:
		m.reporting = false
		if msg.err != nil {
			m.status = ""
			m.errLine = fmt.Sprintf("report failed: %v", msg.err)
			return m, nil
		}
		m.errLine = ""
		m.status = fmt.Sprintf("Report written to %s", msg.path)
		return m, fadeStatusCmd()
	case fadeStatusMsg:
		m.status = ""
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m watchModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyCtrlD:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyEnter, tea.KeyCtrlJ:
		return m.submit()
	case tea.KeyBackspace:
		if m.capturing || m.reporting {
			return m, nil
		}
		r := []rune(m.input)
		if len(r) > 0 {
			m.input = string(r[:len(r)-1])
		}
		return m, nil
	case tea.KeySpace:
		if m.capturing || m.reporting {
			return m, nil
		}
		m.input += " "
		return m, nil
	case tea.KeyRunes:
		if m.capturing || m.reporting {
			return m, nil
		}
		m.input += string(msg.Runes)
		return m, nil
	}
	return m, nil
}

// submit interprets one submitted input line exactly like the plain-text
// watch loop: slash commands, bare Enter (capture, then optional note), and
// bare text as a note capture.
func (m watchModel) submit() (tea.Model, tea.Cmd) {
	line := strings.TrimSpace(m.input)
	m.input = ""
	if m.capturing || m.reporting {
		return m, nil
	}
	if m.awaitNote {
		m.awaitNote = false
		cmd := m.startCapture(line)
		return m, cmd
	}
	m.showHelp = false
	switch {
	case line == "q" || line == "quit" || line == "exit" || line == "/quit" || line == "/q":
		m.quitting = true
		return m, tea.Quit
	case line == "" || line == "/capture":
		m.awaitNote = true
		return m, nil
	case line == "/help":
		m.showHelp = true
		return m, nil
	case line == "/report":
		m.reporting = true
		m.status = ""
		m.errLine = ""
		return m, m.startReport()
	case line == "/note":
		m.errLine = "/note requires text; use `/note <text>`"
		return m, nil
	case strings.HasPrefix(line, "/note "):
		cmd := m.startCapture(strings.TrimSpace(strings.TrimPrefix(line, "/note ")))
		return m, cmd
	case strings.HasPrefix(line, "/"):
		m.errLine = fmt.Sprintf("unknown command %q; type /help for available commands", line)
		return m, nil
	default:
		cmd := m.startCapture(line)
		return m, cmd
	}
}

// startCapture dispatches the existing capture.Service call as a tea.Cmd so
// the UI stays responsive and shows a "Capturing…" state meanwhile. The
// pointer receiver ensures the capturing flag persists with the model.
func (m *watchModel) startCapture(note string) tea.Cmd {
	if m.capturing {
		return nil
	}
	m.capturing = true
	m.status = ""
	m.errLine = ""
	svc, ctx, target, logLines := m.svc, m.ctx, m.target, m.logLines
	return func() tea.Msg {
		sq, err := svc.Capture(ctx, capture.Request{Note: note, Target: target, LogLines: logLines})
		if err != nil {
			return captureDoneMsg{err: err}
		}
		return captureDoneMsg{id: sq.ID}
	}
}

// startReport regenerates the session report in the background, reusing the
// same code path as `squawk report`.
func (m watchModel) startReport() tea.Cmd {
	sessDir := m.store.RootDir()
	return func() tea.Msg {
		path, err := generateReport(sessDir, "", "markdown", "")
		return reportDoneMsg{path: path, err: err}
	}
}

func fadeStatusCmd() tea.Cmd {
	return tea.Tick(statusFadeDelay, func(time.Time) tea.Msg { return fadeStatusMsg{} })
}

func (m watchModel) View() string {
	if m.quitting {
		return ""
	}
	width := m.width
	if width < 24 {
		width = 80
	}

	var b strings.Builder
	b.WriteString(tuiLogoStyle.Render(squawkLogo))
	b.WriteString("\n\n")
	b.WriteString(tuiStatusStyle.Render(m.statusLine()))
	b.WriteString("\n")

	if m.showHelp {
		b.WriteString(tuiHelpStyle.Render(strings.TrimRight(watchHelp, "\n")))
		b.WriteString("\n")
	}
	switch {
	case m.status != "":
		b.WriteString(tuiSuccessStyle.Render(m.status))
		b.WriteString("\n")
	case m.errLine != "":
		b.WriteString(tuiErrorStyle.Render(m.errLine))
		b.WriteString("\n")
	}

	switch {
	case m.capturing:
		b.WriteString(tuiBusyStyle.Render("Capturing…"))
		b.WriteString("\n")
	case m.reporting:
		b.WriteString(tuiBusyStyle.Render("Writing report…"))
		b.WriteString("\n")
	default:
		prompt := ">"
		placeholder := "Type a note, or /help for commands…"
		if m.awaitNote {
			prompt = "note>"
			placeholder = "Optional note (Enter to capture with no note)"
		}
		text := m.input
		if text == "" {
			text = tuiPlaceholderStyle.Render(placeholder)
		}
		line := tuiPromptStyle.Render(prompt+" ") + text + " " + tuiCursorStyle.Render(" ")
		boxWidth := width - 4
		if boxWidth > 78 {
			boxWidth = 78
		}
		if boxWidth < 20 {
			boxWidth = 20
		}
		b.WriteString(tuiBoxStyle.Width(boxWidth).Render(line))
		b.WriteString("\n")
	}

	b.WriteString(tuiFooterStyle.Render("/capture  /note <text>  /report  /help  /q  ·  Enter captures  ·  q quits"))
	return b.String()
}

// statusLine renders the session/backend/target/count line under the logo.
func (m watchModel) statusLine() string {
	count := 0
	backend := "android"
	target := ""
	if m.sess != nil {
		count = len(m.sess.Squawks)
		if m.sess.Backend != "" {
			backend = m.sess.Backend
		}
		target = m.sess.Target
	}
	if m.target != "" {
		target = m.target
	}
	id := ""
	if m.sess != nil {
		id = m.sess.ID
	}
	return fmt.Sprintf("Session %s  ·  Backend %s  ·  Target %s  ·  Squawks %d",
		id, backend, target, count)
}

// runWatchTUI starts the full-screen watch experience on a real terminal.
func runWatchTUI(ctx context.Context, store *session.Store, sess *model.Session, svc capture.Service, target string, logLines int) error {
	p := tea.NewProgram(
		newWatchModel(ctx, store, sess, svc, target, logLines),
		tea.WithAltScreen(),
	)
	final, err := p.Run()
	if err != nil {
		return err
	}
	if m, ok := final.(watchModel); ok && m.quitting {
		fmt.Println("squawk: run `squawk report` to generate a report")
	}
	return nil
}
