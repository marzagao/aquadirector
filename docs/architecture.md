# Architecture

## Project Structure

```
aquadirector/
├── main.go                     # Entry point
├── cmd/                        # Cobra CLI commands (thin layer, no business logic)
├── internal/
│   ├── config/                 # YAML config loading via viper
│   ├── discovery/              # Red Sea network scanner + Eheim mesh discovery
│   ├── alerts/                 # Rule engine + notifiers (stdout, webhook, command)
│   ├── sensor/                 # Kactoily Tuya client + MAC-based IP discovery
│   └── output/                 # Table/JSON/YAML formatting
└── pkg/
    ├── redsea/                 # Red Sea HTTP client library (public, importable)
    ├── eheim/                  # Eheim Digital REST client (public, importable)
    └── tuya/                   # Tuya v3.5 protocol client (public, importable)
```

`pkg/` packages are public and importable by other Go modules. `internal/` packages are private to this module.

## Package Responsibilities

**`cmd/`** — Cobra command definitions. Each subcommand lives in its own file. Commands parse flags, call into `internal/` or `pkg/`, and hand results to `internal/output`. No business logic here.

**`internal/config/`** — Loads `~/.config/aquadirector/aquadirector.yaml` via viper. Exposes a typed `Config` struct. All other packages receive config as a parameter rather than reading it themselves.

**`internal/discovery/`** — Concurrent HTTP subnet scanner for Red Sea devices (probes `/device-info` on port 80 across all hosts in the subnet). Also discovers Eheim devices via the hub's `/api/devicelist` mesh endpoint.

**`internal/alerts/`** — Rule engine that evaluates threshold rules against live device readings. Dispatches firing rules to one or more notifiers (stdout, HTTP webhook, shell command). Rules and notifiers are declared in config.

**`internal/sensor/`** — Thin adapter over `pkg/tuya/`. Maps raw Tuya DPS values to typed `SensorReading` fields. Also implements MAC-based ARP discovery to auto-resolve the sensor's IP when the configured IP fails.

**`internal/output/`** — Formats command output as table, JSON, or YAML based on the `--output` flag. Injects a `run_at` Unix timestamp into JSON/YAML output automatically. Table output appends a `Ran at YYYY-MM-DD HH:MM:SS` footer.

**`pkg/redsea/`** — HTTP client for Red Sea local API (port 80, no auth) and cloud API (OAuth2). Includes `ATOClient`, `LEDClient`, `DeviceInfo`, and `CloudClient`. Retry logic: 5 attempts, 2s delay, 20s timeout, no retry on 400/404.

**`pkg/eheim/`** — HTTP REST client for Eheim Digital hub. HTTP Digest auth (realm `asyncesp`). Hub discovered at `eheimdigital.local` via mDNS; device requests routed by MAC address.

**`pkg/tuya/`** — Pure Go implementation of Tuya local protocol v3.5. AES-128-GCM encryption, 6699 packet framing, session key negotiation. Zero external dependencies, zero Python.

## Key Conventions

- Red Sea devices: local HTTP on port 80, no authentication, JSON payloads
- HTTP retry: 5 attempts, 2s delay, 20s timeout; no retry on 400/404 (matches reference HA integration)
- Tuya v3.5: all packets use 6699 format; handshake encrypted with local key, data with session key; device responses include a 4-byte retcode prefix inside the encrypted payload (strip before parsing)
- Eheim: HTTP Digest auth; POST endpoints require `to` (MAC) in JSON body; GET takes `to` as query param
- Config path follows XDG convention: `~/.config/aquadirector/aquadirector.yaml`
- The Kactoily CLI subcommand is `sensor`; the Eheim subcommand is `feeder`
- Tuya local key rotates on re-pair — use `sensor rekey` to recover
