---
title: tgs login
weight: 10
---

# tgs login

Authenticate with Telegram. The session is saved locally and reused by all subsequent commands.

## Synopsis

{{< snippet "cmd-login/synopsis.md" >}}

## Flags

{{< snippet "cmd-login/flags.md" >}}

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env → `.tgs.yaml` file → `"default"`.

## Examples

{{< snippet "cmd-login/examples.md" >}}

## Auth Methods

{{< snippet "cmd-login/methods.md" >}}

## Output

On successful login:

**JSON (default):**

```json
{"profile":"default","user":{"id":261054642,"phone":"+79001234567","username":"alice","first_name":"Alice","last_name":"Doe"}}
```

**Text (`--output text`):**

```
Logged in as Alice Doe (id: 261054642, profile: default)
```

In the two-step `code` flow, the first call (without `--code`) sends the verification code:

```json
{"profile":"default","status":"code_sent"}
```

```
Verification code sent (profile: default). Re-run with --code to complete login.
```

On failure, the command exits non-zero and prints an error to stderr. Common reasons: invalid code, expired QR token, incorrect 2FA password, or session already present (use `tgs logout` first).

## See Also

- [tgs logout]({{< relref "/reference/commands/auth/logout" >}})
- [Authentication guide]({{< relref "/getting-started/authentication" >}})
