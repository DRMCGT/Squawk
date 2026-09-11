package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// cdpConn is a minimal Chrome DevTools Protocol connection to a single tab.
// It supports request/response calls and dispatching pushed events.
type cdpConn struct {
	ws      *websocket.Conn
	mu      sync.Mutex
	nextID  int
	pending map[int]chan json.RawMessage
}

// dial connects to a tab's webSocketDebuggerUrl and starts the reader
// goroutine that resolves calls and delivers events to onEvent (may be nil).
func dial(ctx context.Context, wsURL string, onEvent func(method string, params json.RawMessage)) (*cdpConn, error) {
	ws, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to Chrome DevTools at %s: %w", wsURL, err)
	}
	c := &cdpConn{ws: ws, pending: make(map[int]chan json.RawMessage)}
	if onEvent == nil {
		onEvent = func(string, json.RawMessage) {}
	}
	go c.readLoop(onEvent)
	return c, nil
}

func (c *cdpConn) close() error { return c.ws.Close() }

// call sends a CDP command and waits for its response, decoding the result
// into out when non-nil.
func (c *cdpConn) call(ctx context.Context, method string, params interface{}, out interface{}) error {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	ch := make(chan json.RawMessage, 1)
	c.pending[id] = ch
	msg := map[string]interface{}{"id": id, "method": method}
	if params != nil {
		msg["params"] = params
	}
	if err := c.ws.WriteJSON(msg); err != nil {
		delete(c.pending, id)
		c.mu.Unlock()
		return fmt.Errorf("send %s: %w", method, err)
	}
	c.mu.Unlock()

	select {
	case raw := <-ch:
		var resp struct {
			Error *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return fmt.Errorf("parse response for %s: %w", method, err)
		}
		if resp.Error != nil {
			return fmt.Errorf("%s: %s (code %d)", method, resp.Error.Message, resp.Error.Code)
		}
		if out != nil && len(resp.Result) > 0 {
			if err := json.Unmarshal(resp.Result, out); err != nil {
				return fmt.Errorf("decode result for %s: %w", method, err)
			}
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// readLoop reads messages until the connection closes: responses resolve
// pending calls, events are forwarded to onEvent.
func (c *cdpConn) readLoop(onEvent func(string, json.RawMessage)) {
	for {
		var msg json.RawMessage
		if err := c.ws.ReadJSON(&msg); err != nil {
			return
		}
		var head struct {
			ID     *int            `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(msg, &head); err != nil {
			continue
		}
		if head.ID != nil {
			c.mu.Lock()
			ch, ok := c.pending[*head.ID]
			delete(c.pending, *head.ID)
			c.mu.Unlock()
			if ok {
				ch <- msg
			}
			continue
		}
		if head.Method != "" {
			onEvent(head.Method, head.Params)
		}
	}
}
