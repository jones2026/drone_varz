// Package pluginenv converts Drone plugin settings into the PLUGIN_*
// environment variables a Drone plugin container expects, per
// https://docs.drone.io/pipeline/docker/syntax/plugins/#plugin-inputs:
// settings are passed as env vars, uppercased and prefixed PLUGIN_, with
// array values joined into comma-separated strings.
package pluginenv

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnvVarName converts a settings key into the environment variable name a
// plugin looks for: upper-cased, with any character invalid in a shell
// variable name (anything but [A-Za-z0-9_]) replaced by an underscore.
func EnvVarName(key string) string {
	var b strings.Builder
	b.WriteString("PLUGIN_")
	for _, r := range strings.ToUpper(key) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// Build parses a Drone-style `settings:` block (the same YAML a user would
// paste from a .drone.yml pipeline step) into PLUGIN_* environment
// variables. Scalars are kept as their literal YAML text rather than
// decoded into Go types (int/float/bool) and reformatted — Drone's own
// example ("1.0" staying "1.0" rather than becoming "1") only round-trips
// if the original text is preserved, since re-encoding a decoded float64
// loses trailing zeros. Arrays of scalars are joined with commas; anything
// else (nested mappings, non-scalar array elements) is reported as an
// error rather than silently mis-encoded.
func Build(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}, nil
	}

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("parse settings YAML: %w", err)
	}
	if len(doc.Content) == 0 {
		return map[string]string{}, nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("settings must be a YAML mapping of key: value pairs")
	}

	vars := make(map[string]string, len(root.Content)/2)
	var errs []string
	for i := 0; i+1 < len(root.Content); i += 2 {
		keyNode, valNode := root.Content[i], root.Content[i+1]
		s, err := scalarOrList(valNode)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", keyNode.Value, err))
			continue
		}
		vars[EnvVarName(keyNode.Value)] = s
	}
	if len(errs) > 0 {
		sort.Strings(errs)
		return nil, fmt.Errorf("unsupported plugin setting value(s): %s", strings.Join(errs, "; "))
	}
	return vars, nil
}

func scalarOrList(n *yaml.Node) (string, error) {
	switch n.Kind {
	case yaml.ScalarNode:
		if n.Tag == "!!null" {
			return "", nil
		}
		return n.Value, nil
	case yaml.SequenceNode:
		parts := make([]string, len(n.Content))
		for i, item := range n.Content {
			s, err := scalarOrList(item)
			if err != nil {
				return "", fmt.Errorf("array element %d: %w", i, err)
			}
			parts[i] = s
		}
		return strings.Join(parts, ","), nil
	default:
		return "", fmt.Errorf("unsupported YAML value (expected a scalar or an array of scalars)")
	}
}
