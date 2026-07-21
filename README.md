# drone_varz

Run a [Drone](https://www.drone.io/) plugin — just a Docker image — as a
step in a GitHub Actions workflow, without modifying the plugin.

Drone plugins expect two things from their environment: a `DRONE_*`
variable for every piece of build/repo/commit context, and a `PLUGIN_*`
variable for every key in the pipeline step's `settings:` block. This
action derives both from the ambient GitHub Actions context and a
Drone-style `settings:` input, then runs the plugin image with the
checked-out repo mounted as its working directory — the same contract
Drone itself provides.

## Usage

```yaml
jobs:
  notify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: jones2026/drone_varz@main
        with:
          image: plugins/slack
          settings: |
            channel: developers
            username: drone
```

The `settings:` value is the exact YAML you'd paste under a Drone
pipeline step — no translation needed. See `.github/workflows/example.yml`
for a runnable example using the real, published [plugins/webhook] plugin,
pointed at a local [go-httpbin] sidecar over `localhost` rather than an
external endpoint, so the example never sends a request off the runner.

[plugins/webhook]: https://github.com/drone-plugins/drone-webhook
[go-httpbin]: https://github.com/mccutchen/go-httpbin

### Inputs

| Input      | Required | Default          | Description                                                     |
|------------|----------|-------------------|-------------------------------------------------------------------|
| `image`    | yes      | —                 | Drone plugin Docker image to run, e.g. `plugins/docker`.        |
| `settings` | no       | `''`              | Plugin settings as YAML, matching Drone's `settings:` block.    |
| `pull`     | no       | `if-not-present`  | Image pull policy: `always`, `if-not-present`, or `never`.      |

## How settings become `PLUGIN_*` vars

Each key in `settings:` is upper-cased and prefixed `PLUGIN_`; arrays are
joined into comma-separated strings, per Drone's own [plugin input
convention](https://docs.drone.io/pipeline/docker/syntax/plugins/#plugin-inputs).
Scalars are kept as their literal YAML text (not decoded and reformatted),
so e.g. a `1.0` tag doesn't silently become `1`.

```yaml
settings:
  username: kevinbacon
  tags:
    - 1.0.0
    - 1.0
```

becomes:

```
PLUGIN_USERNAME=kevinbacon
PLUGIN_TAGS=1.0.0,1.0
```

Nested mappings aren't supported (Drone plugin settings are documented as
primitives or arrays of primitives) and are rejected with an error.

## How `DRONE_*` vars are derived

Most `DRONE_*` variables are filled in from `GITHUB_*` env vars and the
`GITHUB_EVENT_PATH` payload — branch/commit/PR/tag/repo info, and semver
fields parsed from a tag ref. A few have no reliable GitHub Actions
source and are always exported empty:

`DRONE_CALVER`, `DRONE_DEPLOY_TO`, `DRONE_FAILED_STAGES`,
`DRONE_FAILED_STEPS`, `DRONE_SEMVER_ERROR`, `DRONE_STAGE_DEPENDS_ON`,
`DRONE_STAGE_VARIANT`, `DRONE_STEP_NAME`, `DRONE_SYSTEM_VERSION`.

A further set that Drone types as integers - `DRONE_BUILD_FINISHED`,
`DRONE_STAGE_FINISHED`, `DRONE_BUILD_PARENT`, `DRONE_STEP_NUMBER` - default
to `"0"` rather than `""` for the same reason: many official Drone plugins
bind these straight to an `int64` CLI flag and crash on `""` before their
own logic runs. `DRONE_BUILD_STARTED`, `DRONE_BUILD_CREATED`, and
`DRONE_STAGE_STARTED` get a real, if approximate, value - the time this
export step ran, since GitHub Actions exposes no job-start-time env var.
`DRONE_PULL_REQUEST` defaults to `"0"` on non-PR builds, matching Drone's
own default, rather than being left empty.

(These mostly describe Drone-server/build-graph state — timing,
promote/rollback lineage, multi-stage dependencies — that GitHub Actions
doesn't expose. A plugin that hard-requires one of these won't work
here.)

## Known limitations

- **Plugin outputs aren't captured.** Drone plugins don't have a
  standard outputs convention the way GitHub Actions steps do, so this
  action doesn't attempt to parse plugin stdout/files into
  `GITHUB_OUTPUT`. If you need a value out of a plugin run, the plugin
  itself must write it somewhere you can read in a later step (e.g. a
  file in the checked-out workspace).
- **Requires Docker on the runner.** Works out of the box on
  GitHub-hosted Linux runners; self-hosted or non-Linux runners need
  Docker available.
- **Cold-compiles on every run.** The action builds its own Go binaries
  via `go run` on each invocation rather than shipping a prebuilt
  binary, so there's Go-toolchain setup + compile overhead per step.

## Development

```sh
go build ./...
go vet ./...
go test ./... -cover
```

`internal/droneenv` derives `DRONE_*` vars from a GitHub Actions
environment, `internal/pluginenv` derives `PLUGIN_*` vars from a settings
YAML block, and `internal/githubenv` writes either into `GITHUB_ENV`.
`scripts/run-plugin.sh` does the actual `docker run` against the target
plugin image.
