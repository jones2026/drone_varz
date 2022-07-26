package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/suzuki-shunsuke/go-ci-env/v3/cienv"
)

type droneVarz struct {
	DRONE_BRANCH               string
	DRONE_BUILD_ACTION         string
	DRONE_BUILD_CREATED        string
	DRONE_BUILD_EVENT          string
	DRONE_BUILD_FINISHED       string
	DRONE_BUILD_LINK           string
	DRONE_BUILD_NUMBER         string
	DRONE_BUILD_PARENT         string
	DRONE_BUILD_STARTED        string
	DRONE_BUILD_STATUS         string
	DRONE_CALVER               string
	DRONE_COMMIT               string
	DRONE_COMMIT_AFTER         string
	DRONE_COMMIT_AUTHOR        string
	DRONE_COMMIT_AUTHOR_AVATAR string
	DRONE_COMMIT_AUTHOR_EMAIL  string
	DRONE_COMMIT_AUTHOR_NAME   string
	DRONE_COMMIT_BEFORE        string
	DRONE_COMMIT_BRANCH        string
	DRONE_COMMIT_LINK          string
	DRONE_COMMIT_MESSAGE       string
	DRONE_COMMIT_REF           string
	DRONE_COMMIT_SHA           string
	DRONE_DEPLOY_TO            string
	DRONE_FAILED_STAGES        string
	DRONE_FAILED_STEPS         string
	DRONE_GIT_HTTP_URL         string
	DRONE_GIT_SSH_URL          string
	DRONE_PULL_REQUEST         string
	DRONE_PULL_REQUEST_TITLE   string
	DRONE_REMOTE_URL           string
	DRONE_REPO                 string
	DRONE_REPO_BRANCH          string
	DRONE_REPO_LINK            string
	DRONE_REPO_NAME            string
	DRONE_REPO_NAMESPACE       string
	DRONE_REPO_OWNER           string
	DRONE_REPO_PRIVATE         string
	DRONE_REPO_SCM             string
	DRONE_REPO_VISIBILITY      string
	DRONE_SEMVER               string
	DRONE_SEMVER_BUILD         string
	DRONE_SEMVER_ERROR         string
	DRONE_SEMVER_MAJOR         string
	DRONE_SEMVER_MINOR         string
	DRONE_SEMVER_PATCH         string
	DRONE_SEMVER_PRERELEASE    string
	DRONE_SEMVER_SHORT         string
	DRONE_SOURCE_BRANCH        string
	DRONE_STAGE_ARCH           string
	DRONE_STAGE_DEPENDS_ON     string
	DRONE_STAGE_FINISHED       string
	DRONE_STAGE_KIND           string
	DRONE_STAGE_MACHINE        string
	DRONE_STAGE_NAME           string
	DRONE_STAGE_NUMBER         string
	DRONE_STAGE_OS             string
	DRONE_STAGE_STARTED        string
	DRONE_STAGE_STATUS         string
	DRONE_STAGE_TYPE           string
	DRONE_STAGE_VARIANT        string
	DRONE_STEP_NAME            string
	DRONE_STEP_NUMBER          string
	DRONE_SYSTEM_HOST          string
	DRONE_SYSTEM_HOSTNAME      string
	DRONE_SYSTEM_PROTO         string
	DRONE_SYSTEM_VERSION       string
	DRONE_TAG                  string
	DRONE_TARGET_BRANCH        string
}

func main() {
	varz := fetchEnvironmentVariables()
	for field, val := range varz {
		fmt.Printf("%s=%s\n", field, val)
	}
	godotenv.Write(varz, os.Getenv("GITHUB_ENV"))
}

func fetchEnvironmentVariables() map[string]string {
	client := cienv.NewGitHubActions(nil)
	droneVariables := make(map[string]string)
	droneVariables["CI"] = "true"
	droneVariables["DRONE"] = "true"
	droneVariables["DRONE_BRANCH"] = client.Branch()
	return droneVariables
}
