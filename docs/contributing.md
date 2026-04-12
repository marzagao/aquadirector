# Contributing

Contributions are welcome — bug reports, new device support, protocol improvements, or documentation fixes.

## Development Setup

You need Go 1.25+ and nothing else for building and testing. To validate CI locally before pushing, you'll also need Docker (see [Validate CI locally](#validate-ci-locally) below).

```sh
git clone https://github.com/marzagao/aquadirector.git
cd aquadirector
go build ./...   # verify it compiles
go test ./...    # run all tests
```

## Build & Test

```sh
make build        # build binary with version info (./aquadirector)
make install      # install to $GOPATH/bin
make test         # run all tests with -v
make lint         # run go vet

go build ./...              # build all packages (no binary)
go test ./pkg/redsea/ -v    # test a specific package
go test ./... -run TestName # run a specific test by name
```

Tests cover 8 packages: `alerts`, `config`, `discovery`, `output`, `sensor`, `redsea`, `tuya`, `eheim`.

**Testing without real hardware:** all existing tests use static fixtures — no live devices required. When adding a new feature, capture a real device response once and write a table-driven test against it. This keeps CI reliable regardless of network access.

## Code Organization

- **`cmd/`** — Thin CLI layer only. Commands parse flags and call into `internal/` or `pkg/`. No business logic, no device protocol code.
- **`internal/`** — Application logic private to this module.
- **`pkg/`** — Public, importable libraries (`pkg/redsea`, `pkg/eheim`, `pkg/tuya`). These can be used by other Go modules independently of the CLI.

## Adding a New Device

1. **Client library** — Add `pkg/<device>/` with a typed Go client. Follow the pattern in `pkg/redsea/` or `pkg/eheim/`: a `Client` struct, constructor, and methods that return typed structs.
2. **Config** — Add a config section in `internal/config/config.go` and a corresponding entry in `aquadirector.yaml.example`.
3. **CLI commands** — Add `cmd/<device>.go` with Cobra subcommands. Keep commands thin: parse flags, call the client, pass results to `internal/output`.
4. **Discovery** — If the device is discoverable on the local network, add discovery logic in `internal/discovery/`.
5. **Alerts** — If the device exposes metrics, add a case in `internal/alerts/` so its readings can be used in alert rules.
6. **Tests** — Add unit tests in `pkg/<device>/<device>_test.go`. Test response parsing with static fixtures; no live device required.

## Conventions

- HTTP retry: 5 attempts, 2s delay, 20s timeout; skip retry on 400/404
- Tuya v3.5: 6699 framing, AES-128-GCM, 4-byte retcode prefix in responses (strip before parsing)
- Config path: `~/.config/aquadirector/aquadirector.yaml` (XDG convention)
- Output formatting goes through `internal/output` — do not print directly from commands or clients
- `pkg/` packages must not import from `internal/`
- Keep `cmd/` commands concise; extract logic to `internal/` if a command grows beyond flag parsing

## Validate CI locally

Before pushing, run the CI workflow locally with [agent-ci](https://agent-ci.dev/) to catch failures without burning a GitHub Actions run:

```sh
npx @redwoodjs/agent-ci
```

This requires Docker. It runs the same `build`, `test`, and `vet` steps as the GitHub Actions workflow inside a container identical to `ubuntu-latest`. If it passes locally, it will pass in CI.

> **Note:** The Go cache is configured via `go env -w` in the workflow, so it works correctly inside the agent-ci container without any extra setup.

## Submitting Changes

- **Bug fixes and small improvements** — open a PR directly
- **New device support or larger features** — open an issue first to discuss approach
- **PRs should:** include or update tests, not break existing ones, and keep `cmd/` thin

For questions or to share what you've built with the library packages, open a GitHub discussion.
