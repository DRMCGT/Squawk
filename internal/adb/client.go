// Package adb wraps the adb command-line tool. It is only concerned with
// invoking adb and parsing its output; it knows nothing about sessions or
// reports.
package adb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/squawk-dev/squawk/internal/model"
)

// Timeouts for adb operations.
const (
	DevicesTimeout    = 10 * time.Second
	ScreenshotTimeout = 15 * time.Second
	LogcatTimeout     = 10 * time.Second
)

// Client executes adb commands through a Runner.
type Client struct {
	runner Runner
}

// NewClient returns a Client backed by the given Runner.
func NewClient(runner Runner) *Client { return &Client{runner: runner} }

// Check verifies that adb is available by running `adb version`.
func (c *Client) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, DevicesTimeout)
	defer cancel()
	if _, err := c.runner.Run(ctx, "version"); err != nil {
		return fmt.Errorf("adb was not found in PATH: %w", err)
	}
	return nil
}

// Devices lists connected Android devices by parsing `adb devices`.
func (c *Client) Devices(ctx context.Context) ([]model.Device, error) {
	ctx, cancel := context.WithTimeout(ctx, DevicesTimeout)
	defer cancel()
	out, err := c.runner.Run(ctx, "devices")
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	return ParseDevices(string(out))
}

// ParseDevices parses the output of `adb devices`.
func ParseDevices(out string) ([]model.Device, error) {
	var devices []model.Device
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		devices = append(devices, model.Device{Serial: fields[0], State: fields[1]})
	}
	return devices, nil
}

// SelectDevice picks a single usable device. When requested is non-empty it
// must match a connected device exactly; otherwise the caller gets an error
// for zero or multiple devices.
func SelectDevice(devices []model.Device, requested string) (model.Device, error) {
	if requested != "" {
		for _, d := range devices {
			if d.Serial == requested {
				if d.State != "device" {
					return model.Device{}, fmt.Errorf("device %s is %s", d.Serial, d.State)
				}
				return d, nil
			}
		}
		return model.Device{}, fmt.Errorf("device %s not found among connected devices", requested)
	}
	if len(devices) == 0 {
		return model.Device{}, errors.New("no Android devices are connected")
	}
	if len(devices) > 1 {
		return model.Device{}, errors.New("multiple Android devices are connected; specify --device")
	}
	if devices[0].State != "device" {
		return model.Device{}, fmt.Errorf("device %s is %s", devices[0].Serial, devices[0].State)
	}
	return devices[0], nil
}

// Screenshot captures the current screen of the device as PNG bytes.
func (c *Client) Screenshot(ctx context.Context, device string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, ScreenshotTimeout)
	defer cancel()
	out, err := c.runner.Run(ctx, "-s", device, "exec-out", "screencap", "-p")
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}
	if len(out) == 0 {
		return nil, errors.New("screenshot failed: adb returned empty output")
	}
	if !isPNG(out) {
		return nil, errors.New("screenshot failed: adb returned invalid PNG data")
	}
	return out, nil
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}

func isPNG(b []byte) bool {
	if len(b) < len(pngSignature) {
		return false
	}
	return bytes.Equal(b[:len(pngSignature)], pngSignature)
}

// LogcatTail dumps the most recent `lines` logcat entries for the device.
func (c *Client) LogcatTail(ctx context.Context, device string, lines int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, LogcatTimeout)
	defer cancel()
	if lines <= 0 {
		lines = 200
	}
	out, err := c.runner.Run(ctx, "-s", device, "logcat", "-d", "-t", strconv.Itoa(lines))
	if err != nil {
		return nil, fmt.Errorf("logcat failed: %w", err)
	}
	return out, nil
}
