// Package githubenv writes environment variables into a GitHub Actions
// GITHUB_ENV file, for later steps in the same job to pick up.
package githubenv

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Write appends vars to the file at GITHUB_ENV using the heredoc form
// GitHub Actions requires for values that may contain newlines:
//
//	NAME<<DELIMITER
//	value
//	DELIMITER
//
// Keys are written in sorted order for deterministic output. If GITHUB_ENV
// is unset (e.g. running locally, outside a workflow), this is a no-op.
func Write(vars map[string]string) error {
	path := os.Getenv("GITHUB_ENV")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open GITHUB_ENV: %w", err)
	}
	defer f.Close()
	return write(f, vars)
}

func write(w io.Writer, vars map[string]string) error {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		delim, err := delimiterFor(vars[k])
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s<<%s\n%s\n%s\n", k, delim, vars[k], delim); err != nil {
			return fmt.Errorf("write %s: %w", k, err)
		}
	}
	return nil
}

// delimiterFor returns a random heredoc delimiter that does not collide
// with the given value, per GitHub's guidance for multiline environment
// file values.
func delimiterFor(value string) (string, error) {
	for {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("generate delimiter: %w", err)
		}
		delim := "ghadelim_" + hex.EncodeToString(b)
		if !strings.Contains(value, delim) {
			return delim, nil
		}
	}
}
