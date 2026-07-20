#!/usr/bin/env bash
# Runs a Drone plugin Docker image as if it were a Drone pipeline step:
# the checked-out repo is mounted as a volume with the container's working
# directory set to its root (Drone plugins expect to find and operate on
# the source tree there, never clone it themselves), and every DRONE_*,
# PLUGIN_*, CI and DRONE env var already present in this shell (exported
# earlier via GITHUB_ENV by the drone_varz and settings-to-env steps) is
# forwarded into the container.
#
# Runs with --network host (Linux-only, which is what GitHub-hosted
# runners are) so the plugin can reach job-level `services:` sidecars via
# localhost, the same way it would reach a real external endpoint - useful
# for pointing a webhook/notify-style plugin at a local stand-in instead
# of a real one during testing.
set -euo pipefail

image="${1:?usage: run-plugin.sh <image> <pull-policy>}"
pull_policy="${2:-if-not-present}"

case "$pull_policy" in
  always)
    docker pull "$image"
    ;;
  if-not-present)
    docker image inspect "$image" >/dev/null 2>&1 || docker pull "$image"
    ;;
  never) ;;
  *)
    echo "run-plugin.sh: unknown pull policy '$pull_policy' (expected always, if-not-present, or never)" >&2
    exit 1
    ;;
esac

env_args=()
while IFS='=' read -r name _; do
  case "$name" in
    DRONE_*|PLUGIN_*|CI|DRONE)
      env_args+=(-e "$name")
      ;;
  esac
done < <(env)

exec docker run --rm \
  --network host \
  --workdir "$GITHUB_WORKSPACE" \
  -v "$GITHUB_WORKSPACE:$GITHUB_WORKSPACE" \
  "${env_args[@]}" \
  "$image"
