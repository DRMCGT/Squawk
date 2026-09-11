package adb

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/squawk-dev/squawk/internal/model"
)

type fakeRunner struct {
	out   []byte
	err   error
	calls [][]string
}

func (f *fakeRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	return f.out, f.err
}

func TestParseDevices(t *testing.T) {
	out := `List of devices attached
emulator-5554	device
emulator-5556	offline
PHONE123	unauthorized

`
	devices, err := ParseDevices(out)
	if err != nil {
		t.Fatalf("ParseDevices: %v", err)
	}
	want := []model.Device{
		{Serial: "emulator-5554", State: "device"},
		{Serial: "emulator-5556", State: "offline"},
		{Serial: "PHONE123", State: "unauthorized"},
	}
	if !reflect.DeepEqual(devices, want) {
		t.Fatalf("got %+v, want %+v", devices, want)
	}
}

func TestParseDevicesEmpty(t *testing.T) {
	devices, err := ParseDevices("List of devices attached\n\n")
	if err != nil {
		t.Fatalf("ParseDevices: %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("expected no devices, got %+v", devices)
	}
}

func TestSelectDevice(t *testing.T) {
	devices := []model.Device{
		{Serial: "emulator-5554", State: "device"},
		{Serial: "emulator-5556", State: "device"},
	}

	if _, err := SelectDevice(nil, ""); err == nil {
		t.Fatal("expected error for zero devices")
	}
	if _, err := SelectDevice(devices, ""); err == nil {
		t.Fatal("expected error for multiple devices")
	}
	if _, err := SelectDevice(devices, "emulator-5556"); err != nil {
		t.Fatalf("expected selection by serial, got %v", err)
	}
	if _, err := SelectDevice(devices, "missing"); err == nil {
		t.Fatal("expected error for unknown serial")
	}
	offline := []model.Device{{Serial: "emulator-5554", State: "unauthorized"}}
	if _, err := SelectDevice(offline, "emulator-5554"); err == nil {
		t.Fatal("expected error for unauthorized device")
	}
}

func TestScreenshotValidPNG(t *testing.T) {
	r := &fakeRunner{out: append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, []byte("...")...)}
	c := NewClient(r)
	got, err := c.Screenshot(context.Background(), "emulator-5554")
	if err != nil {
		t.Fatalf("Screenshot: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected screenshot bytes")
	}
	want := []string{"-s", "emulator-5554", "exec-out", "screencap", "-p"}
	if !reflect.DeepEqual(r.calls[0], want) {
		t.Fatalf("got call %v, want %v", r.calls[0], want)
	}
}

func TestScreenshotEmptyOutput(t *testing.T) {
	c := NewClient(&fakeRunner{out: nil})
	if _, err := c.Screenshot(context.Background(), "emulator-5554"); err == nil {
		t.Fatal("expected error for empty screenshot output")
	}
}

func TestScreenshotInvalidPNG(t *testing.T) {
	c := NewClient(&fakeRunner{out: []byte("not a png")})
	if _, err := c.Screenshot(context.Background(), "emulator-5554"); err == nil {
		t.Fatal("expected error for invalid PNG")
	}
}

func TestScreenshotRunnerError(t *testing.T) {
	c := NewClient(&fakeRunner{err: errors.New("adb: device not found")})
	if _, err := c.Screenshot(context.Background(), "emulator-5554"); err == nil {
		t.Fatal("expected error propagated")
	}
}

func TestLogcatTail(t *testing.T) {
	r := &fakeRunner{out: []byte("09-10 14:32:01 E MyApp: boom\n")}
	c := NewClient(r)
	out, err := c.LogcatTail(context.Background(), "emulator-5554", 150)
	if err != nil {
		t.Fatalf("LogcatTail: %v", err)
	}
	if string(out) == "" {
		t.Fatal("expected logcat output")
	}
	want := []string{"-s", "emulator-5554", "logcat", "-d", "-t", "150"}
	if !reflect.DeepEqual(r.calls[0], want) {
		t.Fatalf("got call %v, want %v", r.calls[0], want)
	}
}
