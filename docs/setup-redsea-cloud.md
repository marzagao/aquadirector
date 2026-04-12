# Red Sea Cloud Setup (Optional)

Enabling the Red Sea cloud integration adds two sections to the dashboard:

- **Notifications** — last 7 days of in-app alerts from the ReefBeat app
- **Temp 7d** — daily min/avg/max water temperature from the ATO's built-in probe

## What You Need

- A Red Sea / ReefBeat account
- The app's OAuth2 `client_credentials` (see below)

## Step 1: Capture the Client Credentials

The `client_credentials` value is the Base64-encoded `client_id:client_secret` from the ReefBeat mobile app. It's the same for all users and never changes — you only need to capture it once.

To get it, proxy the app's HTTPS traffic while logging in:

1. Install [mitmproxy](https://mitmproxy.org) (or Charles Proxy, Proxyman, etc.) on your computer
2. Configure your phone to route traffic through the proxy
3. Open the ReefBeat app and log in
4. Find the `POST /oauth/token` request
5. Copy the value of the `Authorization: Basic` header — that's your `client_credentials`

## Step 2: Add to Config

```yaml
cloud:
  username: "your@email.com"
  password: "yourpassword"
  client_credentials: "base64encodedclientidandsecret"
```

The token is fetched automatically on first use and cached at `~/.config/aquadirector/cloud_token.json`. It auto-refreshes using the refresh token — you won't need to log in again unless you change your password.

## Step 3: Verify

```sh
aquadirector dashboard
```

The `=== Notifications ===` and `Temp 7d` rows should now appear.
