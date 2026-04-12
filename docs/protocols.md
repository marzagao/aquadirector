# Protocol & API Reference

## Kactoily 7-in-1 Sensor

The sensor communicates using Tuya local protocol v3.5 (AES-128-GCM encrypted). MAC OUI `3C:0B:59` = Tuya Smart Inc.

### DPS Mapping

| DPS | Field | Unit | Scale |
|-----|-------|------|-------|
| dp1 | TDS | ppm | raw |
| dp2 | Temperature | °C | ÷10 |
| dp7 | Battery | % | raw |
| dp10 | pH | — | ÷100 |
| dp11 | EC | µS/cm | raw |
| dp12 | ORP | mV | raw |
| dp101 | Screen Light | bool | — |
| dp102 | Salinity | % | ÷100 |
| dp103 | Specific Gravity | — | ÷1000 |
| dp113 | Temp unit | "f"/"c" | — |
| dp129 | Temp (display unit) | °F when dp113="f" | ÷10 |

## Red Sea Local API

The Red Sea client (`pkg/redsea/`) talks to devices over local HTTP on port 80 with no authentication. Based on the [ha-reefbeat-component](https://github.com/Elwinmage/ha-reefbeat-component) Home Assistant integration.

### ReefATO+ Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/dashboard` | Full ATO status |
| GET | `/configuration` | ATO configuration |
| POST | `/mode` | Set mode (auto/manual) |
| POST | `/resume` | Clear empty state |
| POST | `/update-volume` | Set reservoir volume |

### RSLED60 (G2) Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/manual` | Current manual settings |
| POST | `/manual` | Set manual intensity |
| POST | `/timer` | Set timer override |
| GET | `/auto/{day}` | Get schedule for day (1–7) |
| POST | `/auto/{day}` | Set schedule for day |
| POST | `/mode` | Set mode (auto/manual/timer) |
| GET | `/acclimation` | Acclimation status |
| GET | `/moonphase` | Moon phase |

### Common Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/device-info` | Model, firmware, MAC, HWID |
| GET | `/firmware` | Firmware version |
| GET | `/wifi` | Wi-Fi signal info |
| GET | `/description.xml` | UPnP device description |

**Retry policy:** 5 attempts, 2s delay, 20s timeout. No retry on 400 or 404.

## Red Sea Cloud API

The cloud client (`pkg/redsea/CloudClient`) authenticates against `cloud.thereefbeat.com` using OAuth2 password grant. Tokens are cached at `~/.config/aquadirector/cloud_token.json` and auto-refreshed using the refresh token.

### Authentication

```
POST /oauth/token
Authorization: Basic <client_credentials>
Content-Type: application/x-www-form-urlencoded

grant_type=password&username=<email>&password=<password>
```

`client_credentials` is the Base64-encoded `client_id:client_secret` from the ReefBeat mobile app. It is the same for all users and can be captured by proxying the app's HTTPS traffic (e.g. with mitmproxy) — look for the `Authorization: Basic` header on the login request.

### Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/notification/inapp?expirationDays=7&page=0&size=N&sortDirection=DESC` | In-app notifications |
| GET | `/reef-ato/{hwid}/temperature-log?duration=P7D` | 7-day temperature history |

**Temperature log response:** array of day objects, each with 15-minute interval averages. The ATO device's `hwid` is resolved at runtime from the local `/device-info` endpoint (or set via `hwid:` in config to skip the lookup).

## Eheim Digital API

The Eheim client (`pkg/eheim/`) communicates via local HTTP REST with Digest authentication (realm `asyncesp`, default: user=`api`, pass=`admin`). Based on the [official API docs](https://api.eheimdigital.com/docs/eheim_digital_api/general).

The hub (`eheimdigital.local`) uses a mesh network — all requests target devices by MAC address (query param `to` for GET, JSON body field `to` for POST).

### Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/userdata` | Device identity (name, version, revision) |
| GET | `/api/devicelist` | Mesh device list with IPs |
| GET | `/api/autofeeder?to=MAC` | Feeder status (FEEDER_DATA) |
| POST | `/api/autofeeder/feed` | Trigger one manual feeding |
| POST | `/api/autofeeder/full` | Mark drum as refilled |
| POST | `/api/autofeeder/bio` | Set schedule + feeding break |
| POST | `/api/autofeeder/config` | Set overfeeding + filter sync |
| POST | `/api/brightness` | Set status LED brightness (0–100) |

### FEEDER_DATA Fields

| Field | Type | Description |
|-------|------|-------------|
| weight | float | Food weight in grams |
| isSpinning | int (0/1) | Drum currently spinning |
| level | [int, int] | [level_value, drum_state] |
| configuration | [][][]int | Weekly feeding schedule |
| overfeeding | int (0/1) | Overfeeding protection enabled |
| feedingBreak | int (0/1) | Random fasting day per week |
| breakDay | int (0/1) | Is today a break day (read-only) |
| sync | string (MAC) | Paired filter MAC (reduces flow before feeding) |

### Drum States

| Value | State |
|-------|-------|
| 0 | GREEN (full) |
| 1 | ORANGE (mid) |
| 2 | RED (low) |
| 5 | MEASURING |

### Schedule Format

`configuration` is a 7-element array (Mon–Sun). Each day is `[[time_minutes...], [turns...]]`.

- Times are minutes since midnight (e.g. 480 = 08:00, 1200 = 20:00)
- Turns are drum rotations per feeding
- Empty `[[], []]` = no feeding (fasting day)
- Maximum 2 slots per day

Example: `[[[480, 1200], [2, 1]], [[], []], ...]` = Monday at 08:00 (2 turns) and 20:00 (1 turn), Tuesday fasting.
