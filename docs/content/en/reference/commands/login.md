---
title: tgs login
weight: 10
---

# tgs login

Authenticate with Telegram. The session is saved locally and reused by all subsequent commands.

## Synopsis

```
tgs login [flags]
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | `-T` | `desktop` | Auth method: `desktop`, `code`, or `qr` |
| `--desktop-dir` | `-d` | (auto-detect) | Path to Telegram Desktop tdata directory |
| `--passcode` | | | Telegram Desktop local passcode |
| `--phone` | | | Phone number for `code` method |
| `--profile` | `-p` | (resolved) | Profile name to authenticate into |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env → `.tgs.yaml` file → `"default"`.

## Examples

```bash
# Import session from Telegram Desktop (default)
tgs login

# Custom tdata path
tgs login --desktop-dir ~/AppData/Roaming/Telegram\ Desktop/tdata

# Passcode-protected Telegram Desktop
tgs login --passcode mysecret

# Phone + verification code (interactive)
tgs login --type code

# Phone + code with phone pre-filled
tgs login --type code --phone +1234567890

# QR code login
tgs login --type qr

# Login to a specific profile
tgs login --profile work

# Login to a profile with a specific method
tgs login --type code --profile work --phone +0987654321
```

## Auth Methods

| Method | Description |
|--------|-------------|
| `desktop` | Import session from Telegram Desktop. No phone or code needed. |
| `code` | Phone number + SMS/Telegram verification code. Supports 2FA. |
| `qr` | Display QR code in terminal. Scan with Telegram on another device. Supports 2FA. |

## See Also

- [tgs logout](/reference/commands/logout/)
- [Authentication guide](/getting-started/authentication/)
