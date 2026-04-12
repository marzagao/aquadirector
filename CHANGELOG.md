# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/marzagao/aquadirector/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/marzagao/aquadirector/releases/tag/v1.0.0
