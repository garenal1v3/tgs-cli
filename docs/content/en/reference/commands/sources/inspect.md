---
title: tgs sources inspect
weight: 20
---

# tgs sources inspect

Fetch detailed information for a single Telegram source. Works for sources you are subscribed to as well as public channels you are not a member of.

## Usage

```
tgs sources inspect <ref> [flags]
```

The `ref` argument identifies the target source. Supported formats:

| Format | Example | Description |
|---|---|---|
| `@username` | `@durov` | Public username with `@` prefix |
| `username` | `durov` | Public username without prefix |
| `id:<n>` | `id:-1001234567890` | Telegram peer ID (recommended for negative IDs — pflag eats bare `-N` as a flag) |
| Numeric ID | `12345` | Positive peer ID. For negative IDs use the `id:` form above, or insert `--` (e.g. `inspect -- -1001234567890`) |
| Phone number | `+79001234567` | For contacts and your own account |
| `-` or `@me` | `-` | Saved Messages (your personal chat) |

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--no-stats` | | bool | `false` | Skip expensive stats (total/24h/first messages) |
| `--no-cache` | | bool | `false` | Disable peer and stats caches |
| `--max-wait` | | int | `60` | Max seconds to wait on FLOOD_WAIT |
| `--profile` | `-p` | string | | Account profile name |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env -> `.tgs.yaml` file -> `"default"`.

## Examples

Inspect a channel by username:

```bash
tgs sources inspect @durov
```

Inspect a group by numeric ID (use the `id:` form to avoid the CLI flag parser):

```bash
tgs sources inspect id:-1009876543210
# or with a literal `--` separator (note: command flags must come BEFORE `--`,
# since anything after the separator is treated as positional):
tgs sources inspect --no-stats -- -1009876543210
```

Inspect your Saved Messages:

```bash
tgs sources inspect -
```

Inspect a channel you are not subscribed to:

```bash
tgs sources inspect @somechannel
```

Skip expensive stats:

```bash
tgs sources inspect @durov --no-stats
```

Use a specific account profile:

```bash
tgs sources inspect @golang --profile work
```

## Output

Returns a single JSON object with the full source shape. The `subscribed` field indicates whether your account is a member.

```json
{
  "id": -1001234567890,
  "type": "channel",
  "title": "Durov's Channel",
  "username": "durov",
  "access": "public",
  "members_count": 1234567,
  "has_comments": true,
  "linked_chat_id": -1009876543210,
  "verified": true,
  "unread_count": 0,
  "last_message": {"id": 4321, "date": "2026-05-28T08:15:00Z"},
  "subscribed": false,
  "description": "Pavel Durov's official channel.",
  "creation_date": "2015-08-26T10:00:00Z",
  "invite_link": "https://t.me/+abc",
  "stats": {
    "total_messages": 12345,
    "messages_24h": 3,
    "first_message": {"id": 1, "date": "2015-08-26T10:00:00Z"}
  }
}
```

For unsubscribed public channels, `subscribed` is `false` and `stats` is omitted from the JSON entirely. Display fields (`title`, `username`, `access`, `members_count`, `description`, `verified`, …) are still populated from the public channel info.

### Additional fields (inspect only)

| Field | Type | Notes |
|---|---|---|
| `subscribed` | bool | Whether your account is subscribed to this source |
| `description` | string | Full description / about text |
| `creation_date` | string | RFC3339 UTC timestamp of when the chat/channel was created. Only set for channels/supergroups/groups — Telegram does not expose user registration dates |
| `invite_link` | string | Primary invite link; present when available |

Boolean fields (`verified`, `scam`, `fake`, `restricted`, `archived`, `pinned`, `saved`, `deleted`, `has_topics`, `gigagroup`) appear only when `true`.

### Flag notes

`--no-cache` disables the peer cache (the BoltDB stored at `$XDG_DATA_HOME/tgs/profiles/<profile>/cache.db`, defaulting to `~/.local/share/tgs/profiles/<profile>/cache.db` on Linux/macOS) so that `@username` resolves and any cached display snapshot are bypassed. The companion `stats_cache.db` next to it backs `tgs sources list --with-stats` — `inspect` does not use it.

## See Also

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}})
