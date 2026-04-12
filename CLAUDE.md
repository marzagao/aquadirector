# Aquadirector

Go CLI for home aquarium automation. Monitors and controls Red Sea devices (ReefATO+, RSLED60), a Kactoily 7-in-1 water sensor, and an Eheim autofeeder+. Optionally integrates with the Red Sea cloud API for notifications and temperature history.

## Build & Test

```sh
go build ./...            # build all packages
go test ./... -v          # run all tests (75 tests across 8 packages)
make build                # build binary with version info
gofmt -w .               # format all Go files (required before every commit)
```

Always run `gofmt -w .` before committing. Go Report Card grades gofmt compliance and CI will catch it if skipped.

## Architecture

- `cmd/` — Cobra CLI commands (thin layer, no business logic)
- `internal/config/` — YAML config via viper, loaded from `~/.config/aquadirector/aquadirector.yaml`
- `internal/discovery/` — Subnet scanning for Red Sea devices (concurrent HTTP probes) + Eheim mesh discovery
- `internal/alerts/` — Rule engine: evaluates threshold rules, dispatches to stdout/webhook/command notifiers
- `internal/sensor/` — Kactoily sensor: DPS-to-field mapping, calls `pkg/tuya/` for protocol
- `pkg/tuya/` — Pure Go Tuya v3.5 local protocol client (AES-128-GCM, session negotiation)
- `internal/output/` — Table/JSON/YAML formatting
- `pkg/redsea/` — Public Go client for Red Sea local HTTP API (retry logic, device discovery, ATO + LED clients) + cloud OAuth2 client
- `pkg/eheim/` — Public Go client for Eheim Digital WebSocket API (hub discovery, autofeeder control)

## Key Conventions

- Red Sea devices use local HTTP on port 80, no auth, JSON payloads
- HTTP retry: 5 attempts, 2s delay, 20s timeout, no retry on 400/404 (matches reference implementation)
- Kactoily sensor uses Tuya protocol v3.5 (AES-128-GCM encrypted). Native Go implementation in `pkg/tuya/`
- Tuya v3.5: ALL packets use 6699 format. Handshake encrypted with local key, data with session key
- Tuya v3.5 device responses include a 4-byte retcode prefix inside the encrypted payload (must strip before parsing)
- Config: `~/.config/aquadirector/aquadirector.yaml` (XDG convention)
- `pkg/redsea/` is public/importable; `internal/` is private
- CLI subcommand for the Kactoily device is `sensor` (not `kactoily`)
- Kactoily sensor MUST be registered in Smart Life app (not the Kactoily app) — the Kactoily app uses a private Tuya OEM schema that isn't accessible via the Tuya IoT Developer Platform
- The Tuya local key rotates when the device is re-paired with any app — this breaks the connection. Use `sensor rekey` to fetch the new key.
- Eheim devices use local HTTP REST API with Digest auth (default: user=`api`, pass=`admin`)
- Eheim hub discovered at `eheimdigital.local` via mDNS; mesh network routes to devices by MAC
- Eheim REST endpoints: `/api/autofeeder` (GET status), `/api/autofeeder/feed` (POST), `/api/autofeeder/full` (POST), `/api/autofeeder/bio` (POST schedule+feedingBreak), `/api/autofeeder/config` (POST overfeeding+sync)
- POST endpoints require `to` (MAC) in JSON body; GET takes `to` as query param
- CLI subcommand for the Eheim autofeeder is `feeder`
- `pkg/eheim/` is public/importable

## Red Sea API References

Local API reference implementation cloned at `~/Dropbox/src/ha-reefbeat-component`. Key files:
- `custom_components/redsea/reefbeat/api.py` — HTTP retry logic
- `custom_components/redsea/reefbeat/ato.py` — ATO endpoints
- `custom_components/redsea/reefbeat/led.py` — LED G1/G2 differences
- `custom_components/redsea/const.py` — Model IDs, field names, constants
- `custom_components/redsea/auto_detect.py` — Discovery algorithm

Cloud API reference: [OpenReefBeat](https://github.com/MDamon/OpenReefBeat) — Python client for `cloud.thereefbeat.com`. Key findings:
- OAuth2 password grant to `POST /oauth/token` with `Authorization: Basic <client_credentials>`
- `client_credentials` is a base64-encoded `client_id:client_secret` baked into the ReefBeat mobile app (capture from HTTPS proxy traffic)
- Token cached at `~/.config/aquadirector/cloud_token.json`; auto-refreshed using refresh token
- Notifications: `GET /notification/inapp?expirationDays=7&page=0&size=N&sortDirection=DESC` → `{content: [...]}`
- ATO temperature log: `GET /reef-ato/{hwid}/temperature-log?duration=P7D` → array of `{date, interval, avg[]}` (one entry per day, `avg` contains one reading per `interval` minutes)

## Kactoily Sensor DPS Mapping

MAC OUI `3C:0B:59` = Tuya Smart Inc. Device responds to Tuya v3.5 protocol.

| DPS | Field | Scale | Unit |
|-----|-------|-------|------|
| 1 | TDS | raw | ppm |
| 2 | Temperature | /10 | °C |
| 7 | Battery | raw | % |
| 10 | pH | /100 | - |
| 11 | EC | raw | uS/cm |
| 12 | ORP | raw | mV |
| 101 | Screen Light | bool | - |
| 102 | Salinity | /100 | % |
| 103 | SG | /1000 | - |
| 113 | Temp unit | string | "f"/"c" |
| 129 | Temp (display unit) | /10 | °F when dp113="f" |

## Eheim Autofeeder+ Protocol

Reference implementation: `autinerd/eheimdigital` Python library (GitHub).

### REST API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/userdata` | Device identity (USRDTA): name, version, revision |
| GET | `/api/devicelist` | Mesh device list: clientList, clientIPList |
| GET | `/api/autofeeder?to=MAC` | Feeder status (FEEDER_DATA) |
| POST | `/api/autofeeder/feed` | Trigger one manual feeding (`{to: MAC}`) |
| POST | `/api/autofeeder/full` | Mark drum as refilled (`{to: MAC}`) |
| POST | `/api/autofeeder/bio` | Set schedule + feeding break (`{to, configuration, feedingBreak}`) |
| POST | `/api/autofeeder/config` | Set overfeeding + sync (`{to, overfeeding, sync}`) |
| POST | `/api/brightness` | Set status LED brightness (`{sysLED: 0-100}`) |

All endpoints use HTTP Digest auth (realm `asyncesp`, default user=`api`, pass=`admin`).

### FEEDER_DATA Fields

| Field | Type | Description |
|-------|------|-------------|
| weight | float | Food weight in grams |
| isSpinning | int (0/1) | Drum currently spinning |
| level | [int, int] | [level_value, drum_state] |
| configuration | [][][]int | Weekly feeding schedule (see below) |
| overfeeding | int (0/1) | Overfeeding protection enabled |
| sollRegulation | int (0/1) | Not needed for API (read-only) |
| feedingBreak | int (0/1) | Random fasting day during the week (set via `/bio`) |
| breakDay | int (0/1) | Is today a break day (read-only) |
| sync | string (MAC) | MAC of paired filter (set via `/config`; reduces flow before feeding, restores after 10 min) |
| partnerName | string | Not needed for API (read-only) |

### Drum States

0 = GREEN (full), 1 = ORANGE (mid), 2 = RED (low), 5 = MEASURING

### Schedule Format

7-element array (Mon-Sun). Each day: `[[time_minutes...], [turns...]]`.
- Times are minutes since midnight (e.g., 480 = 08:00, 1200 = 20:00)
- Turns are drum rotations per feeding
- Empty `[[], []]` = no feeding (fasting day)
- Max 2 slots per day

Example: `[[[480,1200],[2,1]], [[],[]], ...]` = Monday at 08:00 (2 turns) and 20:00 (1 turn), Tuesday fasting.

## What's Done

- Full project skeleton with cobra CLI, viper config, table/json/yaml output
- `pkg/redsea/` client library: HTTP client with retry, DeviceInfo, ATO client, LED client
- `pkg/tuya/` native Go Tuya v3.5 client: AES-128-GCM, 6699 framing, session negotiation, zero Python dependency
- `internal/discovery/` subnet scanner
- All CLI commands: dashboard, discover, status, ato (status/resume/volume/mode/config), led (status/manual/timer/mode/schedule), sensor (probe/status/rekey), feeder (status/feed/drum/config/schedule), alerts (check/config)
- Alert engine with configurable rules, threshold evaluation, stdout/webhook/command notifiers
- Kactoily sensor: protocol discovered (Tuya v3.5), DPS mapping complete, live reading works natively in Go
- Red Sea devices tested against real hardware (ReefATO+ and RSLED60)
- ATO temperature reads from ato_sensor.current_read (not a top-level field)
- 75 unit tests passing across 8 packages (alerts, config, discovery, output, sensor, redsea, tuya, eheim)
- `discover --save` writes discovered devices to config file
- `--watch` flag on ato/led/sensor status commands for continuous monitoring
- `sensor rekey` command fetches fresh local key from Tuya Cloud and updates config
- `sensor status` auto-resolves IP via MAC-based ARP discovery if configured IP fails
- Tuya Device.Status() retries 3 times with 1s delay on transient failures
- Config file at `~/.config/aquadirector/aquadirector.yaml` with real device IPs and Tuya credentials
- `pkg/eheim/` native Go Eheim Digital REST client: HTTP Digest auth, hub mesh discovery, autofeeder status/control, schedule management
- Eheim autofeeder+ CLI: status (with --watch), feed, drum (full/tare/measure), config, schedule (set/clear per day)
- `discover` command scans both Red Sea (HTTP subnet) and Eheim (WebSocket mesh) in parallel
- Feeder integrated into dashboard, status, and alert engine (weight, drum_state, level metrics)
- `discover --save` writes feeder host+MAC to config alongside Red Sea devices
- `pkg/redsea/cloud.go` — Red Sea cloud OAuth2 client: token issuance/refresh/caching, `GetNotifications`, `GetATOTemperatureLog`
- Dashboard: notifications (last 7 days from cloud), ATO pump state + last trigger cause, ATO days till empty, LED acclimation status, 7-day temperature range (cloud), pH/ORP status labels, volumes in gallons
- Dashboard only shows cloud sections when `cloud.username` and `cloud.client_credentials` are configured
- Dashboard sections: `=== Water Quality ===`, `=== ATO ===`, `=== Lighting ===`, `=== Feeding ===`, `=== Notifications ===` — each section only shown when data is present
- Feeder weight: negative (tared but empty drum) displays as `~0g`
- All commands print `Ran at YYYY-MM-DD HH:MM:SS` footer in table mode; JSON/YAML output includes `run_at` unix timestamp at the top level (injected by `internal/output` formatter, no call-site changes needed)

## What's Left

Backlog is clear. Potential future work:
- Support additional Red Sea devices (ReefDose, ReefMat, ReefRun, ReefWave)
- Support additional Eheim devices (thermocontrol, classicVARIO, pHcontrol, LEDcontrol)
- Native Go Tuya device control (write DPS values, e.g. toggle screen light)
- Grafana/Prometheus metrics export
