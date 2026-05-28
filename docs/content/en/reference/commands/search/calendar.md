---
title: tgs search calendar
weight: 40
---

# tgs search calendar

Get message search results grouped by date for a specific chat and filter type.

## Usage

```
tgs search calendar [flags]
```

This command takes no positional arguments.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--chat` | `-c` | string | | Chat to query (username, phone, or ID). **Required.** |
| `--filter` | | string | | Message type filter (see [filter values](/en/reference/commands/search/messages/#filter-values)). **Required.** |
| `--max-wait` | | int | `60` | Max seconds to wait on FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Disable peer resolution cache |
| `--profile` | `-p` | string | | Account profile name |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env -> `.tgs.yaml` file -> `"default"`.

## Examples

Get a calendar of photo messages in a channel:

```bash
tgs search calendar -c @travel_photos --filter photo
```

Get a calendar of shared documents in a group:

```bash
tgs search calendar -c @dev_team --filter document
```

## Output

Returns JSON with an array of date entries and a `total` count.

**JSON (default):**

```json
{
  "periods": [
    {"date": "2026-05-12", "count": 2, "min_msg_id": 510, "max_msg_id": 511},
    {"date": "2026-05-10", "count": 1, "min_msg_id": 508, "max_msg_id": 508},
    {"date": "2025-12-23", "count": 1, "min_msg_id": 466, "max_msg_id": 466}
  ],
  "total": 98
}
```

**Text (`--output text`):**

```
2026-05-12  2
2026-05-10  1
2025-12-23  1

total: 98
```

Each entry covers messages posted on that date (UTC). `min_msg_id` and `max_msg_id` let you fetch the actual messages via `tgs search messages --cursor` or any third-party tool that takes a message ID range.

## See Also

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}})
- [tgs search global]({{< relref "/reference/commands/search/global" >}})
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}})
