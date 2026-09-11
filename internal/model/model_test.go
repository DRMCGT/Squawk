package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSessionJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 10, 14, 30, 12, 0, time.UTC)
	in := &Session{
		ID:        "20260910-143012",
		StartedAt: now,
		Device:    "emulator-5554",
		LogLines:  200,
		Squawks: []Squawk{
			{ID: 1, CapturedAt: now.Add(time.Minute), Note: "", Screenshot: "squawks/001/screenshot.png", Logcat: "squawks/001/logcat.txt"},
			{ID: 2, CapturedAt: now.Add(2 * time.Minute), Note: "save button unresponsive", Screenshot: "squawks/002/screenshot.png", Logcat: "squawks/002/logcat.txt"},
		},
	}

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Session
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.ID != in.ID || out.Device != in.Device || out.LogLines != in.LogLines {
		t.Fatalf("round trip mismatch: %+v", out)
	}
	if len(out.Squawks) != 2 {
		t.Fatalf("expected 2 squawks, got %d", len(out.Squawks))
	}
	if out.Squawks[0].Note != "" {
		t.Fatalf("expected empty note preserved, got %q", out.Squawks[0].Note)
	}
	if out.Squawks[1].Note != "save button unresponsive" {
		t.Fatalf("unexpected note: %q", out.Squawks[1].Note)
	}
}
