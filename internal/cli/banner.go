package cli

import (
	"fmt"

	"github.com/DRMCGT/Squawk/internal/version"
)

// banner is printed on `squawk` with no args and at the start of `squawk
// init`. It is kept short so it does not dominate small terminals.
const banner = `             __
          __/  \__   __
         /  \__/  \_/  \
   _  ___/            \_
  / |  |  |  |  |  |  |  \___
  \_|  |  |  |  |  |  |__/
              |  |
              |  |
`

// printBanner writes the airplane artwork, product name, and version.
func printBanner() {
	fmt.Print(banner)
	fmt.Printf("Squawk v%s\n", version.Version)
}
