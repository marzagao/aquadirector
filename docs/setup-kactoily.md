# Kactoily 7-in-1 Sensor Setup

The sensor uses Tuya protocol v3.5 (AES-128-GCM encrypted). You need the device's **ID** and **local key** from Tuya Cloud before aquadirector can talk to it.

## Before You Start: Use Smart Life, Not the Kactoily App

The Kactoily app uses a private OEM schema that isn't accessible through the Tuya IoT Developer Platform. aquadirector needs the Tuya platform to fetch and refresh credentials.

If the device is currently registered in the Kactoily app:
1. Open the Kactoily app and **remove/unpair** the sensor
2. Open Smart Life and add it there (tap **+** > **Auto Scan**)

## Step 1: Add the Sensor to Smart Life

Install the Smart Life app and add the Kactoily sensor:
- [Android](https://play.google.com/store/apps/details?id=com.tuya.smartlife)
- [iOS](https://apps.apple.com/app/smart-life-smart-living/id1115101477)

## Step 2: Create a Tuya IoT Project

1. Create a free account at [iot.tuya.com](https://iot.tuya.com)
2. Go to **Cloud** > **Development** > **Create Cloud Project**
   - Industry: Smart Home
   - Development Method: Custom
   - Data Center: match your region (e.g. Western America, Central Europe)
3. Subscribe to these APIs:
   - **IoT Core**
   - **Authorization Token Management**
   - **Smart Home Device Management**
4. Note your **Access ID** and **Access Secret** from the project overview

## Step 3: Link Your Smart Life Account

In the Tuya IoT console: **Devices** > **Link Tuya App Account** > scan the QR code from the Smart Life app (**Me** tab > scan icon in the top-right corner).

## Step 4: Fetch Device Credentials

```sh
pip install tinytuya
python3 -c "
import tinytuya, json
c = tinytuya.Cloud(apiRegion='us', apiKey='YOUR_ACCESS_ID', apiSecret='YOUR_ACCESS_SECRET')
print(json.dumps(c.getdevices(), indent=2))
"
```

Find the Kactoily device in the output and copy the `id` and `key` values.

## Step 5: Add to Config

```yaml
sensor:
  ip: "192.168.1.15"
  device_id: "your_device_id"
  local_key: "your_local_key"
  version: "3.5"
```

## Step 6: Verify

```sh
aquadirector sensor status
```

---

## Key Rotation

The local key **rotates** whenever the device is re-paired with any app. If the sensor stops responding with error 914 ("Check device key or version"), run:

```sh
aquadirector sensor rekey --client-id YOUR_ACCESS_ID --client-secret YOUR_ACCESS_SECRET
```

This fetches the new key from Tuya Cloud and updates your config automatically.

**To avoid accidental rotation:**
- Do not remove and re-add the sensor in Smart Life
- Do not pair the sensor with the Kactoily app
- Do not factory reset the sensor
- Simply *viewing* data in either app is safe — only re-pairing rotates the key
