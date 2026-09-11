// Package version holds build metadata injected at compile time via
// -ldflags. Development builds use the "dev"/"unknown" defaults.
package version

import "fmt"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// String returns the human-readable version block printed by `squawk version`.
func String() string {
	return fmt.Sprintf("squawk version %s\ncommit: %s\nbuilt: %s", Version, Commit, BuildDate)
}
