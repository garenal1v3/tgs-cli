---
title: Environment Variables
weight: 50
---

# Environment Variables

All tgs environment variables are optional. They override compiled-in defaults or configuration file values.

## Variables

| Variable | Description |
|----------|-------------|
| `TGS_PROFILE` | Active profile name. Overrides `.tgs.yaml` but is overridden by `--profile` flag. |
| `TGS_API_ID` | Telegram API ID. Overrides the value compiled into the binary. |
| `TGS_API_HASH` | Telegram API Hash. Overrides the value compiled into the binary. |
| `TGS_CONFIG_DIR` | Config directory. Default: `~/.config/tgs` (Linux/macOS), `%APPDATA%\tgs` (Windows). |
| `TGS_DATA_DIR` | Data directory for session databases. Default: `~/.local/share/tgs` (Linux/macOS), `%LOCALAPPDATA%\tgs` (Windows). |

## Profile Selection

```bash
# Use the "work" profile for one command
TGS_PROFILE=work tgs whoami

# Use the "work" profile for the whole shell session
export TGS_PROFILE=work
```

See [Profile resolution order](/guide/profiles/#how-profile-resolution-works) for the full priority chain.

## API Credentials

tgs ships with API credentials compiled in. Most users do not need to set these. You may want to override them if you are building tgs from source without credentials, or if you want to use your own Telegram application credentials.

```bash
export TGS_API_ID=12345
export TGS_API_HASH=abcdef1234567890abcdef1234567890
tgs login
```

Obtain API credentials at [my.telegram.org](https://my.telegram.org).

## Custom Directories

```bash
# Store config in a non-default location
export TGS_CONFIG_DIR=/etc/tgs

# Store sessions on a different volume
export TGS_DATA_DIR=/mnt/secure/tgs-data
```
