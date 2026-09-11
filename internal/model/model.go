// Package model defines the core Squawk data structures shared across
// capture, session persistence, and reporting.
package model

import "time"

// Backend identifiers for capture sources.
const (
	BackendAndroid = "android"
	BackendBrowser = "browser"
)

// Squawk is a single captured bug: a screenshot, a recent log excerpt,
// a timestamp, and an optional tester note.
type Squawk struct {
	ID         int       `json:"id"`
	CapturedAt time.Time `json:"captured_at"`
	Note       string    `json:"note"`
	// Screenshot and Logcat are paths relative to the session directory.
	Screenshot string `json:"screenshot"`
	Logcat     string `json:"logcat"`
}

// Session is one testing session. Squawks are stored in capture order.
type Session struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	// Backend is one of BackendAndroid or BackendBrowser.
	Backend string `json:"backend,omitempty"`
	// Target identifies what is being tested: an adb serial for android or a
	// browser tab URL for browser.
	Target string `json:"target,omitempty"`
	// Endpoint is the backend's control endpoint: empty for android, or the
	// Chrome DevTools base URL (e.g. http://localhost:9222) for browser.
	Endpoint string   `json:"endpoint,omitempty"`
	LogLines int      `json:"log_lines"`
	Squawks  []Squawk `json:"squawks"`
}

// Device is a single entry from `adb devices`.
type Device struct {
	Serial string
	State  string
}
