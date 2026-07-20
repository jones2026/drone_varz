package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/jones2026/drone_varz/internal/droneenv"
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

	if err := droneenv.WriteGitHubEnv(vars); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
