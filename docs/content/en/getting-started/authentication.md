---
title: Authentication
weight: 10
---

# Authentication

tgs connects to Telegram through your real user account via MTProto. You need to authenticate once per profile — the session is saved locally and reused on subsequent commands.

## Quick Start

The default method imports your existing session from Telegram Desktop:

{{< snippet "auth/quick-start.md" >}}

Or choose a different method:

{{< snippet "auth/methods-alt.md" >}}

## Login Methods

### Desktop Import (default)

Imports your existing session from Telegram Desktop. tgs auto-detects the tdata directory on macOS, Linux, and Windows.

{{< snippet "auth/desktop.md" >}}

No phone number or code entry required — the session is transferred directly.

### Phone + Code

Interactive login with your phone number. Supports 2FA (cloud password).

{{< snippet "auth/code-login.md" >}}

tgs will prompt for the verification code sent by Telegram. If 2FA is enabled, it will also prompt for the cloud password.

### Non-Interactive Login (for AI agents)

The `code` method supports a fully non-interactive two-step flow. This is the recommended way for AI agents and automation tools.

{{< snippet "auth/code-login.md" >}}

If 2FA is enabled, pass `--password` in step 2.

### QR Code

Displays a QR code in the terminal. Scan it with the Telegram app on another device.

{{< snippet "auth/qr-login.md" >}}

After scanning, tgs completes the authentication automatically. If 2FA is enabled, you will be prompted for the cloud password.

## Login to a Specific Profile

Use `--profile` to authenticate into a named profile. The profile is created automatically if it does not exist.

{{< snippet "auth/profile-login.md" >}}

See [Profiles]({{< relref "/guide/profiles" >}}) for multi-account setup.

## Verifying Login

After logging in, confirm the active account:

{{< snippet "auth/verify.md" >}}

## Logout

{{< snippet "auth/logout.md" >}}

Logout revokes the session on Telegram servers and clears local session data.

## Full Flag Reference

See [tgs login]({{< relref "/reference/commands/login" >}}) and [tgs logout]({{< relref "/reference/commands/logout" >}}) for all flags.
