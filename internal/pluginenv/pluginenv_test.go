package pluginenv

import "testing"

func TestEnvVarName(t *testing.T) {
	cases := map[string]string{
		"username":          "PLUGIN_USERNAME",
		"build_args":        "PLUGIN_BUILD_ARGS",
		"aws-access-key-id": "PLUGIN_AWS_ACCESS_KEY_ID",
		"Repo.Tags":         "PLUGIN_REPO_TAGS",
	}
	for in, want := range cases {
		if got := EnvVarName(in); got != want {
			t.Errorf("EnvVarName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuild_FromDroneDocsExample(t *testing.T) {
	// Mirrors the plugins/docker example straight from the Drone docs,
	// including the "1.0" tag that must NOT lose its trailing zero.
	vars, err := Build(`
username: kevinbacon
password: pa55word
repo: foo/bar
tags:
  - 1.0.0
  - 1.0
`)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	want := map[string]string{
		"PLUGIN_USERNAME": "kevinbacon",
		"PLUGIN_PASSWORD": "pa55word",
		"PLUGIN_REPO":     "foo/bar",
		"PLUGIN_TAGS":     "1.0.0,1.0",
	}
	for k, v := range want {
		if got := vars[k]; got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if len(vars) != len(want) {
		t.Errorf("got %d vars, want %d: %v", len(vars), len(want), vars)
	}
}

func TestBuild_ScalarTypesPreserveLiteralText(t *testing.T) {
	vars, err := Build(`
enabled: true
retries: 3
timeout: 2.50
empty:
quoted: "007"
mixed:
  - a
  - 1
  - true
`)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	want := map[string]string{
		"PLUGIN_ENABLED": "true",
		"PLUGIN_RETRIES": "3",
		"PLUGIN_TIMEOUT": "2.50",
		"PLUGIN_EMPTY":   "",
		"PLUGIN_QUOTED":  "007",
		"PLUGIN_MIXED":   "a,1,true",
	}
	for k, v := range want {
		if got := vars[k]; got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestBuild_RejectsNestedMaps(t *testing.T) {
	if _, err := Build("nested:\n  a: b\n"); err == nil {
		t.Fatal("expected an error for a nested map setting, got nil")
	}
}

func TestBuild_RejectsNonMappingRoot(t *testing.T) {
	if _, err := Build("- a\n- b\n"); err == nil {
		t.Fatal("expected an error for a non-mapping root, got nil")
	}
}

func TestBuild_Empty(t *testing.T) {
	vars, err := Build("")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(vars) != 0 {
		t.Errorf("expected no vars, got %v", vars)
	}
}
