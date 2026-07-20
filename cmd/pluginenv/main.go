// Command pluginenv reads a Drone-style `settings:` YAML block from a file
// and exports it as PLUGIN_* environment variables into GITHUB_ENV, for a
// later step to pass through to the plugin container.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/jones2026/drone_varz/internal/githubenv"
	"github.com/jones2026/drone_varz/internal/pluginenv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: pluginenv <settings-yaml-file>")
		os.Exit(2)
	}

	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	vars, err := pluginenv.Build(string(raw))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

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
