// Package droneenv derives Drone-compatible DRONE_* environment variables
// from the ambient GitHub Actions environment and event payload, so that a
// Drone plugin (which expects a Drone-shaped environment) can run unmodified
// as a step in a GitHub Actions workflow.
package droneenv

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Env is the subset of the environment this package reads from, as an
// interface so tests can supply a fake.
type Env interface {
	Getenv(key string) string
}

type osEnv struct{}

func (osEnv) Getenv(key string) string { return os.Getenv(key) }

// OSEnv reads from the real process environment and GITHUB_EVENT_PATH file.
var OSEnv Env = osEnv{}

// eventPayload is the subset of the GitHub Actions event JSON
// (GITHUB_EVENT_PATH) this package uses. Shape varies by event type; unused
// fields are simply left zero-valued.
type eventPayload struct {
	Action     string `json:"action"`
	Before     string `json:"before"`
	After      string `json:"after"`
	HeadCommit struct {
		Message string `json:"message"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
	} `json:"head_commit"`
	PullRequest struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"pull_request"`
	Repository struct {
		Private       bool   `json:"private"`
		Visibility    string `json:"visibility"`
		DefaultBranch string `json:"default_branch"`
	} `json:"repository"`
	Sender struct {
		AvatarURL string `json:"avatar_url"`
	} `json:"sender"`
}

func readEventPayload(env Env) eventPayload {
	var p eventPayload
	path := env.Getenv("GITHUB_EVENT_PATH")
	if path == "" {
		return p
	}
	f, err := os.Open(path)
	if err != nil {
		return p
	}
	defer f.Close()
	// Best-effort: an unreadable/malformed payload just yields zero values
	// rather than failing the whole export.
	_ = json.NewDecoder(f).Decode(&p)
	return p
}

var semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-.]+))?(?:\+([0-9A-Za-z-.]+))?$`)

// UnsupportedFields lists DRONE_* variables that have no reliable GitHub
// Actions source and are always exported empty. Drone populates these from
// its own server/build-graph state, which GitHub Actions doesn't expose.
var UnsupportedFields = []string{
	"DRONE_BUILD_CREATED", "DRONE_BUILD_FINISHED", "DRONE_BUILD_PARENT", "DRONE_BUILD_STARTED",
	"DRONE_CALVER", "DRONE_DEPLOY_TO",
	"DRONE_FAILED_STAGES", "DRONE_FAILED_STEPS",
	"DRONE_SEMVER_ERROR",
	"DRONE_STAGE_DEPENDS_ON", "DRONE_STAGE_FINISHED", "DRONE_STAGE_STARTED", "DRONE_STAGE_VARIANT",
	"DRONE_STEP_NAME", "DRONE_STEP_NUMBER",
	"DRONE_SYSTEM_VERSION",
}

// Build derives the full set of DRONE_* variables (plus CI/DRONE) from the
// given environment. Values with no GitHub Actions equivalent are set to ""
// (see UnsupportedFields) rather than omitted, so a plugin doing a plain
// os.Getenv lookup never panics on a missing key.
func Build(env Env) map[string]string {
	g := env.Getenv
	payload := readEventPayload(env)

	repo := g("GITHUB_REPOSITORY")
	owner := g("GITHUB_REPOSITORY_OWNER")
	repoName := strings.TrimPrefix(repo, owner+"/")
	eventName := g("GITHUB_EVENT_NAME")
	refType := g("GITHUB_REF_TYPE")
	refName := g("GITHUB_REF_NAME")

	serverURL := g("GITHUB_SERVER_URL")
	if serverURL == "" {
		serverURL = "https://github.com"
	}
	systemHost := strings.TrimPrefix(strings.TrimPrefix(serverURL, "https://"), "http://")
	if u, err := url.Parse(serverURL); err == nil && u.Host != "" {
		systemHost = u.Host
	}

	isPR := eventName == "pull_request" || eventName == "pull_request_target"

	branch, sourceBranch, targetBranch := refName, refName, ""
	if isPR {
		branch = g("GITHUB_BASE_REF")
		sourceBranch = g("GITHUB_HEAD_REF")
		targetBranch = g("GITHUB_BASE_REF")
	}

	commitAuthor := payload.HeadCommit.Author.Username
	if commitAuthor == "" {
		commitAuthor = g("GITHUB_ACTOR")
	}

	commitMessage := payload.HeadCommit.Message
	if commitMessage == "" && isPR {
		commitMessage = payload.PullRequest.Title
	}

	pullRequest, pullRequestTitle := "", ""
	if isPR {
		if payload.PullRequest.Number != 0 {
			pullRequest = strconv.Itoa(payload.PullRequest.Number)
		}
		pullRequestTitle = payload.PullRequest.Title
	}

	buildEvent, ok := map[string]string{
		"push":                "push",
		"pull_request":        "pull_request",
		"pull_request_target": "pull_request",
		"release":             "tag",
		"schedule":            "cron",
		"workflow_dispatch":   "custom",
	}[eventName]
	if !ok {
		buildEvent = eventName
	}

	tag := ""
	if refType == "tag" {
		tag = refName
	}

	semver, semverShort, semverMajor, semverMinor, semverPatch, semverPre, semverBuild := "", "", "", "", "", "", ""
	if m := semverRe.FindStringSubmatch(tag); m != nil {
		semver = strings.TrimPrefix(tag, "v")
		semverMajor, semverMinor, semverPatch, semverPre, semverBuild = m[1], m[2], m[3], m[4], m[5]
		semverShort = fmt.Sprintf("%s.%s.%s", semverMajor, semverMinor, semverPatch)
	}

	vars := map[string]string{
		"CI":    "true",
		"DRONE": "true",

		"DRONE_BRANCH":        branch,
		"DRONE_SOURCE_BRANCH": sourceBranch,
		"DRONE_TARGET_BRANCH": targetBranch,

		"DRONE_BUILD_ACTION": payload.Action,
		"DRONE_BUILD_EVENT":  buildEvent,
		"DRONE_BUILD_LINK":   fmt.Sprintf("%s/%s/actions/runs/%s", serverURL, repo, g("GITHUB_RUN_ID")),
		"DRONE_BUILD_NUMBER": g("GITHUB_RUN_NUMBER"),
		"DRONE_BUILD_STATUS": "success",

		"DRONE_COMMIT":               g("GITHUB_SHA"),
		"DRONE_COMMIT_SHA":           g("GITHUB_SHA"),
		"DRONE_COMMIT_BEFORE":        payload.Before,
		"DRONE_COMMIT_AFTER":         payload.After,
		"DRONE_COMMIT_REF":           g("GITHUB_REF"),
		"DRONE_COMMIT_BRANCH":        branch,
		"DRONE_COMMIT_LINK":          fmt.Sprintf("%s/%s/commit/%s", serverURL, repo, g("GITHUB_SHA")),
		"DRONE_COMMIT_MESSAGE":       commitMessage,
		"DRONE_COMMIT_AUTHOR":        commitAuthor,
		"DRONE_COMMIT_AUTHOR_NAME":   payload.HeadCommit.Author.Name,
		"DRONE_COMMIT_AUTHOR_EMAIL":  payload.HeadCommit.Author.Email,
		"DRONE_COMMIT_AUTHOR_AVATAR": payload.Sender.AvatarURL,

		"DRONE_PULL_REQUEST":       pullRequest,
		"DRONE_PULL_REQUEST_TITLE": pullRequestTitle,

		"DRONE_GIT_HTTP_URL": fmt.Sprintf("%s/%s.git", serverURL, repo),
		"DRONE_GIT_SSH_URL":  fmt.Sprintf("git@%s:%s.git", systemHost, repo),
		"DRONE_REMOTE_URL":   fmt.Sprintf("%s/%s.git", serverURL, repo),

		"DRONE_REPO":            repo,
		"DRONE_REPO_NAME":       repoName,
		"DRONE_REPO_NAMESPACE":  owner,
		"DRONE_REPO_OWNER":      owner,
		"DRONE_REPO_LINK":       fmt.Sprintf("%s/%s", serverURL, repo),
		"DRONE_REPO_BRANCH":     payload.Repository.DefaultBranch,
		"DRONE_REPO_SCM":        "git",
		"DRONE_REPO_PRIVATE":    strconv.FormatBool(payload.Repository.Private),
		"DRONE_REPO_VISIBILITY": payload.Repository.Visibility,

		"DRONE_TAG":               tag,
		"DRONE_SEMVER":            semver,
		"DRONE_SEMVER_SHORT":      semverShort,
		"DRONE_SEMVER_MAJOR":      semverMajor,
		"DRONE_SEMVER_MINOR":      semverMinor,
		"DRONE_SEMVER_PATCH":      semverPatch,
		"DRONE_SEMVER_PRERELEASE": semverPre,
		"DRONE_SEMVER_BUILD":      semverBuild,

		"DRONE_STAGE_KIND":    "pipeline",
		"DRONE_STAGE_TYPE":    "docker",
		"DRONE_STAGE_NAME":    g("GITHUB_JOB"),
		"DRONE_STAGE_NUMBER":  g("GITHUB_RUN_ATTEMPT"),
		"DRONE_STAGE_OS":      strings.ToLower(g("RUNNER_OS")),
		"DRONE_STAGE_ARCH":    strings.ToLower(g("RUNNER_ARCH")),
		"DRONE_STAGE_MACHINE": g("RUNNER_NAME"),
		"DRONE_STAGE_STATUS":  "success",

		"DRONE_SYSTEM_HOST":     systemHost,
		"DRONE_SYSTEM_HOSTNAME": systemHost,
		"DRONE_SYSTEM_PROTO":    "https",
	}

	for _, k := range UnsupportedFields {
		vars[k] = ""
	}

	return vars
}
