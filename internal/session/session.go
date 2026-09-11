// Package session manages Squawk session directories and the active-session
// pointer file. It is the source of truth for locating the current session
// across separate CLI processes.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DRMCGT/Squawk/internal/model"
)

// PointerFileName is the file that records the active session ID. It contains
// a single session ID followed by a newline.
const PointerFileName = "current-session"

// DefaultRootName is the default Squawk data directory name.
const DefaultRootName = ".squawk"

// ErrNoActiveSession is returned when no usable active session can be found.
var ErrNoActiveSession = errors.New("no active Squawk session found; run `squawk init`")

// Store locates and persists sessions under a root directory
// (default: <cwd>/.squawk).
type Store struct {
	rootDir string
}

// NewStore returns a Store rooted at rootDir. An empty rootDir resolves to
// <current working directory>/.squawk.
func NewStore(rootDir string) (*Store, error) {
	if rootDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("cannot determine current directory: %w", err)
		}
		rootDir = filepath.Join(cwd, DefaultRootName)
	}
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve session dir %q: %w", rootDir, err)
	}
	return &Store{rootDir: abs}, nil
}

// RootDir returns the absolute path to the Squawk data directory.
func (s *Store) RootDir() string { return s.rootDir }

// SessionsDir returns the directory that holds all sessions.
func (s *Store) SessionsDir() string { return filepath.Join(s.rootDir, "sessions") }

// SessionDir returns the directory for a single session.
func (s *Store) SessionDir(id string) string { return filepath.Join(s.SessionsDir(), id) }

// SessionFile returns the path to a session's session.json.
func (s *Store) SessionFile(id string) string { return filepath.Join(s.SessionDir(id), "session.json") }

// PointerFile returns the path to the active-session pointer file.
func (s *Store) PointerFile() string { return filepath.Join(s.rootDir, PointerFileName) }

// SquawkDir returns the directory holding a squawk's artifacts.
func (s *Store) SquawkDir(sessionID string, squawkID int) string {
	return filepath.Join(s.SessionDir(sessionID), "squawks", fmt.Sprintf("%03d", squawkID))
}

// CreateSession creates a new session directory, writes session.json, and
// atomically points the active-session pointer at it.
func (s *Store) CreateSession(backend, target, endpoint string, logLines int) (*model.Session, error) {
	if backend == "" {
		backend = model.BackendAndroid
	}
	if logLines <= 0 {
		logLines = 200
	}
	now := time.Now().UTC()
	base := now.Format("20060102-150405")
	id := base
	for n := 1; ; n++ {
		if _, err := os.Stat(s.SessionDir(id)); os.IsNotExist(err) {
			break
		}
		id = fmt.Sprintf("%s-%d", base, n)
	}
	sess := &model.Session{
		ID:        id,
		StartedAt: now,
		Backend:   backend,
		Target:    target,
		Endpoint:  endpoint,
		LogLines:  logLines,
		Squawks:   []model.Squawk{},
	}
	if err := os.MkdirAll(s.SessionDir(id), 0o755); err != nil {
		return nil, fmt.Errorf("cannot create session directory: %w", err)
	}
	if err := s.SaveSession(sess); err != nil {
		return nil, err
	}
	if err := s.SetCurrent(id); err != nil {
		return nil, err
	}
	return sess, nil
}

// LoadSession reads and validates a session by ID.
func (s *Store) LoadSession(id string) (*model.Session, error) {
	if _, err := os.Stat(s.SessionDir(id)); err != nil {
		return nil, ErrNoActiveSession
	}
	data, err := os.ReadFile(s.SessionFile(id))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", s.SessionFile(id), err)
	}
	var sess model.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("session %s has an invalid session.json: %w", id, err)
	}
	return &sess, nil
}

// SaveSession persists a session atomically (write temp file, then rename).
func (s *Store) SaveSession(sess *model.Session) error {
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot serialize session: %w", err)
	}
	data = append(data, '\n')
	tmp := s.SessionFile(sess.ID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("cannot write session: %w", err)
	}
	if err := os.Rename(tmp, s.SessionFile(sess.ID)); err != nil {
		return fmt.Errorf("cannot finalize session file: %w", err)
	}
	return nil
}

// SetCurrent atomically points the active-session pointer at id.
func (s *Store) SetCurrent(id string) error {
	if _, err := os.Stat(s.SessionDir(id)); err != nil {
		return ErrNoActiveSession
	}
	if err := os.MkdirAll(s.rootDir, 0o755); err != nil {
		return fmt.Errorf("cannot create %s: %w", s.rootDir, err)
	}
	tmp := s.PointerFile() + ".tmp"
	if err := os.WriteFile(tmp, []byte(id+"\n"), 0o644); err != nil {
		return fmt.Errorf("cannot write active session pointer: %w", err)
	}
	if err := os.Rename(tmp, s.PointerFile()); err != nil {
		return fmt.Errorf("cannot finalize active session pointer: %w", err)
	}
	return nil
}

// CurrentSession resolves the active session via the pointer file. It returns
// ErrNoActiveSession when the pointer is missing, empty, or references a
// session directory that no longer exists.
func (s *Store) CurrentSession() (*model.Session, error) {
	data, err := os.ReadFile(s.PointerFile())
	if err != nil {
		return nil, ErrNoActiveSession
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return nil, ErrNoActiveSession
	}
	return s.LoadSession(id)
}

// NextSquawkID returns the ID the next squawk in a session should receive.
func NextSquawkID(sess *model.Session) int {
	return len(sess.Squawks) + 1
}
