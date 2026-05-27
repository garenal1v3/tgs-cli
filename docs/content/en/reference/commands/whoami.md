---
title: tgs whoami
weight: 25
---

# tgs whoami

Show the current profile name and account information. Reads from local storage — no Telegram connection required.

## Synopsis

```
tgs whoami [flags]
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--profile` | `-p` | (resolved) | Profile to inspect |
| `--output` | `-o` | `json` | Output format: `json`, `text` |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env → `.tgs.yaml` file → `"default"`.

## Examples

```bash
# Show current profile (JSON output)
tgs whoami

# Human-readable output
tgs whoami --output text

# Inspect a specific profile
tgs whoami --profile work

# Inspect a profile with text output
tgs whoami --profile work --output text
```

## Output

**Text output (`--output text`):**

```
Profile: work
User:    John Doe
Handle:  @johndoe
Phone:   +1234567890
ID:      123456789
```

**JSON output (default):**

```json
{
  "profile": "work",
  "user": {
    "id": 123456789,
    "phone": "+1234567890",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

If the profile exists but is not logged in, the `user` field will be `null` (JSON) or `"Not logged in"` (text).

## See Also

- [tgs profile](/reference/commands/profile/)
- [Profiles guide](/guide/profiles/)
