package droneenv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type fakeEnv map[string]string

func (f fakeEnv) Getenv(key string) string { return f[key] }

func writeEventPayload(t *testing.T, dir string, payload any) string {
	t.Helper()
	path := filepath.Join(dir, "event.json")
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	return path
}

func TestBuild_Push(t *testing.T) {
	dir := t.TempDir()
	eventPath := writeEventPayload(t, dir, map[string]any{
		"before": "aaa",
		"after":  "bbb",
		"head_commit": map[string]any{
			"message": "fix things",
			"author": map[string]any{
				"name":     "Kevin Bacon",
				"email":    "kevin@example.com",
				"username": "kevinbacon",
			},
		},
		"repository": map[string]any{
			"private":        false,
			"visibility":     "public",
			"default_branch": "main",
		},
	})

	env := fakeEnv{
		"GITHUB_REPOSITORY":       "jones2026/drone_varz",
		"GITHUB_REPOSITORY_OWNER": "jones2026",
		"GITHUB_EVENT_NAME":       "push",
		"GITHUB_REF_TYPE":         "branch",
		"GITHUB_REF_NAME":         "main",
		"GITHUB_REF":              "refs/heads/main",
		"GITHUB_SHA":              "bbb",
		"GITHUB_SERVER_URL":       "https://github.com",
		"GITHUB_RUN_ID":           "123",
		"GITHUB_RUN_NUMBER":       "7",
		"GITHUB_EVENT_PATH":       eventPath,
	}

	vars := Build(env)

	want := map[string]string{
		"DRONE_BRANCH":              "main",
		"DRONE_SOURCE_BRANCH":       "main",
		"DRONE_TARGET_BRANCH":       "",
		"DRONE_BUILD_EVENT":         "push",
		"DRONE_COMMIT":              "bbb",
		"DRONE_COMMIT_BEFORE":       "aaa",
		"DRONE_COMMIT_AFTER":        "bbb",
		"DRONE_COMMIT_MESSAGE":      "fix things",
		"DRONE_COMMIT_AUTHOR":       "kevinbacon",
		"DRONE_COMMIT_AUTHOR_NAME":  "Kevin Bacon",
		"DRONE_COMMIT_AUTHOR_EMAIL": "kevin@example.com",
		"DRONE_REPO":                "jones2026/drone_varz",
		"DRONE_REPO_NAME":           "drone_varz",
		"DRONE_REPO_OWNER":          "jones2026",
		"DRONE_REPO_VISIBILITY":     "public",
		"DRONE_REPO_PRIVATE":        "false",
		"DRONE_PULL_REQUEST":        "",
		"DRONE_BUILD_LINK":          "https://github.com/jones2026/drone_varz/actions/runs/123",
		"DRONE_BUILD_NUMBER":        "7",
	}
	for k, v := range want {
		if got := vars[k]; got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestBuild_PullRequest(t *testing.T) {
	dir := t.TempDir()
	eventPath := writeEventPayload(t, dir, map[string]any{
		"pull_request": map[string]any{
			"number": 42,
			"title":  "Add feature",
		},
	})

	env := fakeEnv{
		"GITHUB_REPOSITORY":       "jones2026/drone_varz",
		"GITHUB_REPOSITORY_OWNER": "jones2026",
		"GITHUB_EVENT_NAME":       "pull_request",
		"GITHUB_BASE_REF":         "main",
		"GITHUB_HEAD_REF":         "feature-branch",
		"GITHUB_SHA":              "ccc",
		"GITHUB_EVENT_PATH":       eventPath,
	}

	vars := Build(env)

	want := map[string]string{
		"DRONE_BRANCH":             "main",
		"DRONE_SOURCE_BRANCH":      "feature-branch",
		"DRONE_TARGET_BRANCH":      "main",
		"DRONE_BUILD_EVENT":        "pull_request",
		"DRONE_PULL_REQUEST":       "42",
		"DRONE_PULL_REQUEST_TITLE": "Add feature",
		"DRONE_COMMIT_MESSAGE":     "Add feature",
	}
	for k, v := range want {
		if got := vars[k]; got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestBuild_Tag_Semver(t *testing.T) {
	env := fakeEnv{
		"GITHUB_REPOSITORY":       "jones2026/drone_varz",
		"GITHUB_REPOSITORY_OWNER": "jones2026",
		"GITHUB_EVENT_NAME":       "push",
		"GITHUB_REF_TYPE":         "tag",
		"GITHUB_REF_NAME":         "v1.2.3-rc1+build5",
	}

	vars := Build(env)

	want := map[string]string{
		"DRONE_TAG":               "v1.2.3-rc1+build5",
		"DRONE_SEMVER":            "1.2.3-rc1+build5",
		"DRONE_SEMVER_SHORT":      "1.2.3",
		"DRONE_SEMVER_MAJOR":      "1",
		"DRONE_SEMVER_MINOR":      "2",
		"DRONE_SEMVER_PATCH":      "3",
		"DRONE_SEMVER_PRERELEASE": "rc1",
		"DRONE_SEMVER_BUILD":      "build5",
	}
	for k, v := range want {
		if got := vars[k]; got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestBuild_UnsupportedFieldsAreEmpty(t *testing.T) {
	vars := Build(fakeEnv{})
	for _, k := range UnsupportedFields {
		if got, ok := vars[k]; !ok || got != "" {
			t.Errorf("%s = %q, %v; want present and empty", k, got, ok)
		}
	}
}
