package adb

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Runner executes adb commands and returns combined stdout/stderr. It is
// implemented by ExecRunner for production and by test doubles in tests.
type Runner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// ExecRunner runs the adb binary, honoring the SQUAWK_ADB environment
// variable (useful for tests and custom adb installs).
type ExecRunner struct {
	adbBin string
}

// NewExecRunner returns an ExecRunner using $SQUAWK_ADB if set, else "adb".
func NewExecRunner() *ExecRunner {
	bin := os.Getenv("SQUAWK_ADB")
	if bin == "" {
		bin = "adb"
	}
	return &ExecRunner{adbBin: bin}
}

// Run executes the adb binary with the given arguments.
func (r *ExecRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.adbBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return out, fmt.Errorf("%s %s: %s", r.adbBin, strings.Join(args, " "), msg)
	}
	return out, nil
}
