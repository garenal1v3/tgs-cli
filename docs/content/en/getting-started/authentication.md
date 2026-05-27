---
title: Authentication
weight: 10
---

# Authentication

tgs connects to Telegram through your real user account via MTProto. You need to authenticate once per profile — the session is saved locally and reused on subsequent commands.

## Quick Start

The default method imports your existing session from Telegram Desktop:

```bash
tgs login
```

Or choose a different method:

```bash
tgs login --type code   # Phone number + SMS/Telegram code
tgs login --type qr     # Scan QR code from another device
```

## Login Methods

### Desktop Import (default)

Imports your existing session from Telegram Desktop. tgs auto-detects the tdata directory on macOS, Linux, and Windows.

```bash
# Auto-detect Telegram Desktop data
tgs login

# Custom tdata path
tgs login --desktop-dir /path/to/tdata

# Passcode-protected Telegram Desktop client
tgs login --passcode SECRET
```

No phone number or code entry required — the session is transferred directly.

### Phone + Code

Interactive login with your phone number. Supports 2FA (cloud password).

```bash
# Interactive — prompts for phone and code
tgs login --type code

# Provide phone upfront
tgs login --type code --phone +1234567890
```

tgs will prompt for the verification code sent by Telegram. If 2FA is enabled, it will also prompt for the cloud password.

### QR Code

Displays a QR code in the terminal. Scan it with the Telegram app on another device.

```bash
tgs login --type qr
```

After scanning, tgs completes the authentication automatically. If 2FA is enabled, you will be prompted for the cloud password.

## Login to a Specific Profile

Use `--profile` to authenticate into a named profile. The profile is created automatically if it does not exist.

```bash
tgs login --profile work
tgs login --type code --profile personal
```

See [Profiles](/guide/profiles/) for multi-account setup.

## Verifying Login

After logging in, confirm the active account:

```bash
tgs whoami
```

## Logout

```bash
tgs logout                  # logout current profile
tgs logout --profile work   # logout specific profile
```

Logout revokes the session on Telegram servers and clears local session data.

## Full Flag Reference

See [tgs login](/reference/commands/login/) and [tgs logout](/reference/commands/logout/) for all flags.
