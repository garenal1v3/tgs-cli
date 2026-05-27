---
title: tgs search global
weight: 11
---

# tgs search global

Search messages globally across all chats, channels, and groups.

## Usage

```
tgs search global [query] [flags]
```

The `query` argument is required and specifies the text to search for.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--channels-only` | | bool | `false` | Search only in channels |
| `--groups-only` | | bool | `false` | Search only in groups |
| `--users-only` | | bool | `false` | Search only in private chats |
| `--folder` | | int | `0` | Search only in folder with this ID |
| `--filter` | | string | | Message type filter (see [filter values](/en/reference/commands/search-messages/#filter-values)) |
| `--after` | | string | | Only messages after date (YYYY-MM-DD or unix timestamp) |
| `--before` | | string | | Only messages before date (YYYY-MM-DD or unix timestamp) |
| `--limit` | `-l` | int | `50` | Max messages to return (1-100) |
| `--cursor` | | string | | Pagination cursor from previous response |
| `--max-wait` | | int | `60` | Max seconds to wait on FLOOD_WAIT |
| `--profile` | `-p` | string | | Account profile name |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env -> `.tgs.yaml` file -> `"default"`.

## Examples

Search across all chats:

```bash
tgs search global "kubernetes"
```

Search only in channels:

```bash
tgs search global "breaking news" --channels-only
```

Search with date filter and limited results:

```bash
tgs search global "quarterly report" --after 2025-04-01 -l 20
```

Search only in a specific folder:

```bash
tgs search global "project update" --folder 3
```

## Output

Returns JSON with an array of matched messages and a `cursor` field for pagination. Each message includes the sender, date, text, chat info, and message ID.

## See Also

- [tgs search messages]({{< relref "/reference/commands/search-messages" >}})
- [tgs search counters]({{< relref "/reference/commands/search-counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}})
