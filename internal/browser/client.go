// Package browser is a Chrome DevTools Protocol backend for Squawk. It
// attaches to a user-launched Chrome (via --remote-debugging-port), lists open
// tabs, and captures screenshots plus console/uncaught-exception logs.
package browser

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBase is the default Chrome DevTools base URL.
const DefaultBase = "http://localhost:9222"

// Client talks to Chrome's DevTools HTTP and WebSocket endpoints.
type Client struct {
	base string
	http *http.Client
}

// Tab is an open page in Chrome.
type Tab struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	WS    string `json:"webSocketDebuggerUrl"`
}

// New returns a Client rooted at base (default http://localhost:9222).
func New(base string) *Client {
	if base == "" {
		base = DefaultBase
	}
	return &Client{base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: 10 * time.Second}}
}

// Check verifies the DevTools endpoint is reachable.
func (c *Client) Check(ctx context.Context) error {
	var v struct {
		Browser string `json:"Browser"`
	}
	if err := c.get(ctx, "/json/version", &v); err != nil {
		return err
	}
	return nil
}

// Tabs lists open page tabs.
func (c *Client) Tabs(ctx context.Context) ([]Tab, error) {
	var targets []struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Title string `json:"title"`
		URL   string `json:"url"`
		WS    string `json:"webSocketDebuggerUrl"`
	}
	if err := c.get(ctx, "/json/list", &targets); err != nil {
		return nil, err
	}
	var tabs []Tab
	for _, t := range targets {
		if t.Type != "page" || t.WS == "" {
			continue
		}
		tabs = append(tabs, Tab{ID: t.ID, Title: t.Title, URL: t.URL, WS: t.WS})
	}
	return tabs, nil
}

// Screenshot captures the tab matching target (URL substring or tab ID) as
// PNG bytes.
func (c *Client) Screenshot(ctx context.Context, target string) ([]byte, error) {
	tab, err := c.findTab(ctx, target)
	if err != nil {
		return nil, err
	}
	conn, err := dial(ctx, tab.WS, nil)
	if err != nil {
		return nil, err
	}
	defer conn.close()

	var res struct {
		Data string `json:"data"`
	}
	if err := conn.call(ctx, "Page.captureScreenshot", map[string]interface{}{"format": "png"}, &res); err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}
	png, err := base64.StdEncoding.DecodeString(res.Data)
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: invalid base64: %w", err)
	}
	if !isPNG(png) {
		return nil, errors.New("screenshot failed: invalid PNG data")
	}
	return png, nil
}

// Logs observes the tab's console and uncaught exceptions for a short window
// and returns up to lines formatted entries.
func (c *Client) Logs(ctx context.Context, target string, lines int) ([]byte, error) {
	tab, err := c.findTab(ctx, target)
	if err != nil {
		return nil, err
	}
	var entries []string
	conn, err := dial(ctx, tab.WS, func(method string, params json.RawMessage) {
		if entry, ok := formatEvent(method, params); ok {
			entries = append(entries, entry)
		}
	})
	if err != nil {
		return nil, err
	}
	defer conn.close()

	if err := conn.call(ctx, "Runtime.enable", nil, nil); err != nil {
		return nil, fmt.Errorf("enable runtime: %w", err)
	}

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}

	if lines <= 0 {
		lines = 200
	}
	if len(entries) > lines {
		entries = entries[len(entries)-lines:]
	}
	return []byte(strings.Join(entries, "\n") + "\n"), nil
}

func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach Chrome DevTools at %s: %w", c.base, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("Chrome DevTools at %s returned %s: %s", c.base, resp.Status, strings.TrimSpace(string(body)))
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("cannot parse DevTools response: %w", err)
		}
	}
	return nil
}

func (c *Client) findTab(ctx context.Context, target string) (Tab, error) {
	tabs, err := c.Tabs(ctx)
	if err != nil {
		return Tab{}, err
	}
	var matches []Tab
	for _, t := range tabs {
		if t.ID == target || strings.Contains(t.URL, target) {
			matches = append(matches, t)
		}
	}
	if len(matches) == 0 {
		return Tab{}, fmt.Errorf("no browser tab matches %q (%d open tabs)", target, len(tabs))
	}
	if len(matches) > 1 {
		return Tab{}, fmt.Errorf("multiple tabs match %q; use a more specific URL", target)
	}
	return matches[0], nil
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}

func isPNG(b []byte) bool {
	if len(b) < len(pngSignature) {
		return false
	}
	return bytes.Equal(b[:len(pngSignature)], pngSignature)
}

// formatEvent renders a CDP console or exception event as a logcat-style line.
func formatEvent(method string, params json.RawMessage) (string, bool) {
	switch method {
	case "Runtime.consoleAPICalled":
		var p struct {
			Type string `json:"type"`
			Args []struct {
				Value       json.RawMessage `json:"value"`
				Description string          `json:"description"`
			} `json:"args"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return "", false
		}
		parts := make([]string, 0, len(p.Args))
		for _, a := range p.Args {
			if len(a.Value) > 0 && string(a.Value) != "null" {
				var s string
				if err := json.Unmarshal(a.Value, &s); err == nil {
					parts = append(parts, s)
				} else {
					parts = append(parts, string(a.Value))
				}
			} else if a.Description != "" {
				parts = append(parts, a.Description)
			}
		}
		if len(parts) == 0 {
			return "", false
		}
		return fmt.Sprintf("[%s] console.%s: %s", time.Now().Format("15:04:05"), p.Type, strings.Join(parts, " ")), true
	case "Runtime.exceptionThrown":
		var p struct {
			ExceptionDetails struct {
				Text      string `json:"text"`
				Exception struct {
					Description string `json:"description"`
				} `json:"exception"`
			} `json:"exceptionDetails"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return "", false
		}
		msg := p.ExceptionDetails.Text
		if p.ExceptionDetails.Exception.Description != "" {
			msg = p.ExceptionDetails.Exception.Description
		}
		if msg == "" {
			return "", false
		}
		return fmt.Sprintf("[%s] error: %s", time.Now().Format("15:04:05"), msg), true
	}
	return "", false
}
