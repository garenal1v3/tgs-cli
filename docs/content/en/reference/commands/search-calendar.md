---
title: tgs search calendar
weight: 13
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
| `--filter` | | string | | Message type filter (see [filter values](/en/reference/commands/search-messages/#filter-values)). **Required.** |
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

Returns JSON with an array of date entries. Each entry contains a date and the count of matching messages on that date, along with a representative message ID.

## See Also

- [tgs search messages]({{< relref "/reference/commands/search-messages" >}})
- [tgs search global]({{< relref "/reference/commands/search-global" >}})
- [tgs search counters]({{< relref "/reference/commands/search-counters" >}})
