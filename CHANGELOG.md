# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.2] - 2026-04-12

### Added

- Colorized `dashboard` output: pH/ORP status labels, drum state, reservoir
  days-till-empty, leak sensor, pump state, and unread notification markers
  are now color-coded for at-a-glance status. Section headers render bold,
  row labels render dim. Palette uses only the 8 base ANSI codes so output
  stays legible on both light and dark terminal themes. Auto-disabled when
  stdout is not a TTY, honors `NO_COLOR`, and overridable with
  `--color=auto|always|never`. JSON/YAML output is unchanged.

## [1.0.1] - 2026-04-12

### Changed

- Example config (`aquadirector.yaml.example`): replaced realistic-looking Tuya credential examples with format descriptions
- Genericized default LAN IPs from `192.168.50.x` to `192.168.1.x` across config defaults, examples, tests, and docs
- Demo recording: switched from GIF (blurry, 256-color) to SVG (razor-sharp vector) via asciinema + svg-term-cli
- Demo: added separator lines between commands for readability

### Fixed

- `gofmt -s` compliance across all packages
- `gocyclo`: reduced complexity of `extractATOField` (21 → 9) and `mockTuyaDevice` (16 → 5) by extracting helper functions
- Fixed misspelling in `DeviceTypeName` (extra `e` in `professional`)

## [1.0.0] - 2026-04-11

### Added

- Red Sea ReefATO+ support: status, mode, volume, resume, configuration
- Red Sea RSLED60 (G2) support: status, manual, timer, mode, schedule
- Kactoily 7-in-1 water sensor: pH, temperature, ORP, salinity, SG, TDS, EC via native Go Tuya v3.5 client
- Eheim autofeeder+ support: status, feed, drum management, schedule, overfeeding protection, filter sync
- Consolidated `dashboard` command with water quality, ATO, lighting, feeding, and cloud notifications
- `discover` command: concurrent subnet scan for Red Sea devices + Eheim mesh discovery, with `--save` to write config
- `alerts` command: threshold rule engine with stdout, webhook, and shell command notifiers
- `sensor rekey` command: fetches updated Tuya local key from cloud and writes to config
- `--watch` flag on status commands for continuous monitoring
- `--output json|yaml` on all commands for scripting and automation
- Pure Go Tuya v3.5 protocol client (`pkg/tuya`): AES-128-GCM, 6699 framing, no Python dependency
- Red Sea cloud integration (optional): 7-day notifications and ATO temperature history via OAuth2
- Public Go libraries: `pkg/redsea`, `pkg/eheim`, `pkg/tuya` importable independently

[1.0.2]: https://github.com/marzagao/aquadirector/releases/tag/v1.0.2
[1.0.1]: https://github.com/marzagao/aquadirector/releases/tag/v1.0.1
[1.0.0]: https://github.com/marzagao/aquadirector/releases/tag/v1.0.0
