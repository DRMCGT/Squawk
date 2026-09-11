package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

// fakeCDP serves /json/version and /json/list, and accepts a single websocket
// that echoes CDP responses and can push events.
type fakeCDP struct {
	tabs   []Tab
	events []map[string]interface{} // events pushed to the ws client

	mu      sync.Mutex
	wsCount int
	ws      *websocket.Conn
	server  *httptest.Server
}

var pngData = base64.StdEncoding.EncodeToString(append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, []byte("img")...))

func newFakeCDP(t *testing.T, events []map[string]interface{}) *fakeCDP {
	t.Helper()
	f := &fakeCDP{tabs: []Tab{
		{ID: "tab1", Title: "My App", URL: "http://localhost:3000/app", WS: "ws://fake/tab1"},
	}, events: events}

	mux := http.NewServeMux()
	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"Browser": "Chrome/fake"})
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		type target struct {
			Type                 string `json:"type"`
			ID                   string `json:"id"`
			Title                string `json:"title"`
			URL                  string `json:"url"`
			WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
		}
		var out []target
		for _, t := range f.tabs {
			out = append(out, target{Type: "page", ID: t.ID, Title: t.Title, URL: t.URL, WebSocketDebuggerURL: t.WS})
		}
		json.NewEncoder(w).Encode(out)
	})
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	mux.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Push events immediately; the client's reader collects them.
		for _, ev := range f.events {
			if err := ws.WriteJSON(ev); err != nil {
				return
			}
		}
		// Handle commands, respond to Page.captureScreenshot with fake data.
		go func() {
			for {
				var msg map[string]json.RawMessage
				if err := ws.ReadJSON(&msg); err != nil {
					return
				}
				var method string
				json.Unmarshal(msg["method"], &method)
				if method == "Page.captureScreenshot" {
					resp := map[string]interface{}{
						"id":     json.RawMessage(msg["id"]),
						"result": map[string]string{"data": pngData},
					}
					if err := ws.WriteJSON(resp); err != nil {
						return
					}
				} else {
					if err := ws.WriteJSON(map[string]interface{}{"id": json.RawMessage(msg["id"]), "result": map[string]interface{}{}}); err != nil {
						return
					}
				}
			}
		}()
	})

	f.server = httptest.NewServer(mux)
	// Rewrite tab WS URLs to point at the fake server, preserving tab identity.
	f.mu.Lock()
	wsBase := strings.Replace(f.server.URL, "http", "ws", 1)
	for i := range f.tabs {
		f.tabs[i].WS = wsBase + "/ws/" + f.tabs[i].ID
	}
	f.mu.Unlock()
	t.Cleanup(f.server.Close)
	return f
}

func TestCheckAndTabs(t *testing.T) {
	f := newFakeCDP(t, nil)
	c := New(f.server.URL)
	if err := c.Check(context.Background()); err != nil {
		t.Fatalf("Check: %v", err)
	}
	tabs, err := c.Tabs(context.Background())
	if err != nil {
		t.Fatalf("Tabs: %v", err)
	}
	if len(tabs) != 1 || tabs[0].ID != "tab1" {
		t.Fatalf("unexpected tabs: %+v", tabs)
	}
}

func TestCheckUnreachable(t *testing.T) {
	c := New("http://127.0.0.1:1")
	if err := c.Check(context.Background()); err == nil {
		t.Fatal("expected error for unreachable CDP endpoint")
	}
}

func TestScreenshot(t *testing.T) {
	f := newFakeCDP(t, nil)
	c := New(f.server.URL)
	png, err := c.Screenshot(context.Background(), "localhost:3000")
	if err != nil {
		t.Fatalf("Screenshot: %v", err)
	}
	if !isPNG(png) {
		t.Fatal("expected PNG bytes")
	}
}

func TestScreenshotNoMatch(t *testing.T) {
	f := newFakeCDP(t, nil)
	c := New(f.server.URL)
	if _, err := c.Screenshot(context.Background(), "nomatch"); err == nil {
		t.Fatal("expected error for unmatched tab")
	}
}

func TestLogsCollectsEvents(t *testing.T) {
	events := []map[string]interface{}{
		{
			"method": "Runtime.consoleAPICalled",
			"params": map[string]interface{}{
				"type": "error",
				"args": []map[string]interface{}{
					{"type": "string", "value": "save failed"},
				},
			},
		},
		{
			"method": "Runtime.exceptionThrown",
			"params": map[string]interface{}{
				"exceptionDetails": map[string]interface{}{
					"text": "Uncaught TypeError: x is null",
				},
			},
		},
	}
	f := newFakeCDP(t, events)
	c := New(f.server.URL)
	out, err := c.Logs(context.Background(), "tab1", 200)
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	body := string(out)
	if !strings.Contains(body, "console.error: save failed") {
		t.Fatalf("missing console event:\n%s", body)
	}
	if !strings.Contains(body, "error: Uncaught TypeError: x is null") {
		t.Fatalf("missing exception event:\n%s", body)
	}
}

func TestLogsRespectsLimit(t *testing.T) {
	events := []map[string]interface{}{}
	for i := 0; i < 5; i++ {
		events = append(events, map[string]interface{}{
			"method": "Runtime.consoleAPICalled",
			"params": map[string]interface{}{
				"type": "log",
				"args": []map[string]interface{}{{"type": "string", "value": fmt.Sprintf("line %d", i)}},
			},
		})
	}
	f := newFakeCDP(t, events)
	c := New(f.server.URL)
	out, err := c.Logs(context.Background(), "tab1", 2)
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	body := string(out)
	if strings.Contains(body, "line 0") || strings.Contains(body, "line 1") || strings.Contains(body, "line 2") {
		t.Fatalf("expected only the last 2 lines:\n%s", body)
	}
	if !strings.Contains(body, "line 3") || !strings.Contains(body, "line 4") {
		t.Fatalf("expected last two lines present:\n%s", body)
	}
}
