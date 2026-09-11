// Package model defines the core Squawk data structures shared across
// capture, session persistence, and reporting.
package model

import "time"

// Squawk is a single captured bug: a screenshot, a recent logcat excerpt,
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
	Device    string    `json:"device,omitempty"`
	LogLines  int       `json:"log_lines"`
	Squawks   []Squawk  `json:"squawks"`
}

// Device is a single entry from `adb devices`.
type Device struct {
	Serial string
	State  string
}
