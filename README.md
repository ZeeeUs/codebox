# codebox

`codebox` is a host-side Go CLI that starts `codex` inside a prepared Docker container.

## How It Works

There are two separate parts in this project:

- `codebox` - a CLI binary you run on the host machine
- `codebox-codex:<version>` - a Docker image used as the runtime environment for `codex`

`codebox` itself does not run inside Docker. It runs on the host and executes `docker build` and `docker run` internally when needed.

## Requirements

You need the following on the host:

- Go
- Docker

## Install

Install `codebox` directly with `go install`:

```bash
go install github.com/ZeeeUs/codebox@latest
```

This installs the `codebox` binary into your Go bin directory, usually one of:

- `$GOBIN`
- `$GOPATH/bin`
- `$HOME/go/bin`

Make sure that directory is in your `PATH`.

Check the installed version with:

```bash
codebox --version
```

## Run

Start `codebox` from the project directory you want to mount into the container:

```bash
codebox
```

You can also run it from source during development:

```bash
go run .
```

## Codebox Updates

Release builds check the latest GitHub Release at startup. When a newer version is available, Codebox prints the current and new versions together with the `go install` update command. Network or GitHub API errors never prevent Codebox from starting.

## First Launch Behavior

When you choose `Codex` or `Shell`, `codebox` checks whether the Docker image `codebox-codex:<version>` exists locally.

Fixed-version images are built when missing and then reused. For `latest`, `codebox` compares the installed Codex version with the latest stable npm release and asks before updating the image. You do not need to run `docker build` manually.

## Override Codex Image Version

You can override the image tag with `--codex-version`:

```bash
codebox --codex-version latest
```

This makes `codebox` use the image:

```bash
codebox-codex:latest
```

When an update is available for `latest`, `codebox` asks whether to install it. A fixed version, for example `--codex-version 0.1.0`, is built once and reused.

## Menu

When started, the application prints:

```text
Codebox starting...
```

Then it shows an interactive menu:

```text
1. Run Codex
2. Mount directory
3. Install skill
4. List skills
5. Shell
6. Configure languages
7. Codex permissions
0. Exit
```

### Menu Actions

- `1` - run `codex` inside the container
- `2` - add an extra mount and save it into `.codebox.yaml`
- `3` - install a skill from a local directory into `~/.codebox/skills/`
- `4` - list installed skills
- `5` - open `bash` inside the container
- `6` - select Go, Rust, or both for the runtime image
- `7` - select the Codex approval and sandbox profile
- `0` - exit

## Config File

`codebox` looks for `.codebox.yaml` in the current working directory.

If the file does not exist, default config is used.

### Default Values

- `codex.version = latest`
- `project.mountPath = /workspace`

### Language toolchains

Language dependencies are opt-in. Use menu option `6` to select Go, Rust, or both. Codebox saves the selection to `.codebox.yaml` and updates the runtime when Codex or Shell starts. Each language has its own Docker layer, so unchanged toolchains are restored from the build cache.

### Example

```yaml
agent: default
codex:
  version: latest
  approvalPolicy: on-request
  sandboxMode: workspace-write
project:
  mountPath: /workspace
languages:
  - go
  - rust
mounts:
  - source: ~/work/shared
    target: /mnt/shared
    mode: rw
skills:
  - name: example-skill
```

## Codex permissions

Use menu option `7` to choose one of three profiles:

- Ask when needed with workspace-write sandbox.
- No prompts with workspace-write sandbox; disallowed operations fail automatically.
- No prompts with full container access.

Full container access also applies to project and additional host directories mounted into the container. Codex can modify or delete files in those mounts without confirmation.

## Docker Runtime Behavior

The container is started similarly to:

```bash
docker run --rm -it \
  -v "$PWD:/workspace:rw" \
  -w /workspace \
  codebox-codex:<version> \
  codex
```

For shell mode:

```bash
docker run --rm -it \
  -v "$PWD:/workspace:rw" \
  -w /workspace \
  codebox-codex:<version> \
  bash
```

## Mounts

The main project directory is always mounted as:

```text
$PWD:/workspace:rw
```

Additional mounts are created as:

```text
host_path:/mnt/<basename>:rw
```

Current behavior:

- `~` is expanded
- source paths are validated
- mount mode must be `ro` or `rw`
- target conflicts are blocked in the interactive mount flow

After adding a mount through the menu, `codebox` saves `.codebox.yaml` and exits so the container can be restarted with the new configuration.

## Skills

Installed skills are stored in:

```text
~/.codebox/skills/
```

Installing a skill copies a local directory into:

```text
~/.codebox/skills/<name>
```

## Runtime dependency versions

Runtime dependencies are pinned and updated together with Codebox releases. Codex is the exception: when `codex.version` is `latest`, Codebox still checks npm for a newer Codex version at launch and asks before rebuilding.

## Tools Available Inside the Container

The runtime image currently installs:

- Go toolchain (available as `go` in `PATH`)
- `gopls`
- `dlv`
- `staticcheck`
- `golangci-lint`
- Rust
- `git`, `curl`, `bash`
- `codex`

`codex` is installed in the container from the latest stable npm release via `npm i -g @openai/codex@latest`.

## Current Limitations

- The host machine must have Go installed to install or run `codebox`
- The host machine must have Docker available
- The project currently does not include a `Makefile`
- The first container launch can take time because the runtime image may need to be built
