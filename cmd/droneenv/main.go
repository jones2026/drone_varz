// Command droneenv exports DRONE_* environment variables, derived from the
// ambient GitHub Actions context, into GITHUB_ENV for later steps to use.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/jones2026/drone-action-bridge/internal/droneenv"
	"github.com/jones2026/drone-action-bridge/internal/githubenv"
)

func main() {
	vars := droneenv.Build(droneenv.OSEnv)

	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s=%s\n", k, vars[k])
	}

	if err := githubenv.Write(vars); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
