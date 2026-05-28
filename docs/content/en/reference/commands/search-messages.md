---
title: tgs search messages
weight: 10
---

# tgs search messages

Search messages within one or more chats, channels, or groups.

## Usage

```
tgs search messages [query] [flags]
```

The `query` argument is required and specifies the text to search for.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--chat` | `-c` | string[] | | Chat to search (username, phone, ID; repeatable, comma-separated). **Required.** |
| `--from` | `-f` | string | | Filter by sender (username, phone, or ID) |
| `--filter` | | string | | Message type filter (see [filter values](#filter-values)) |
| `--after` | | string | | Only messages after date (YYYY-MM-DD or unix timestamp) |
| `--before` | | string | | Only messages before date (YYYY-MM-DD or unix timestamp) |
| `--topic` | | int | `0` | Forum topic ID |
| `--limit` | `-l` | int | `50` | Max messages to return (1-100) |
| `--cursor` | | string | | Pagination cursor from previous response |
| `--max-wait` | | int | `60` | Max seconds to wait on FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Disable peer resolution cache |
| `--include-comments` | | bool | `false` | Also search the linked discussion group of each channel |
| `--profile` | `-p` | string | | Account profile name |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env -> `.tgs.yaml` file -> `"default"`.

### Filter values

`photo`, `video`, `photo-video`, `document`, `url`, `gif`, `voice`, `music`, `round-video`, `geo`, `contact`, `pinned`, `mention`, `phone-call`, `chat-photo`.

## Examples

Search for a keyword in a channel:

```bash
tgs search messages "release notes" -c @golang
```

Search across multiple chats:

```bash
tgs search messages "deployment" -c @devops_team -c @infrastructure
```

Filter by sender and message type:

```bash
tgs search messages "config" -c @mygroup --from @alice --filter document
```

Search within a date range:

```bash
tgs search messages "outage" -c @incidents --after 2025-01-01 --before 2025-06-01
```

Paginate through results:

```bash
tgs search messages "bug" -c @dev -l 10 --cursor "eyJvIjo1MCwiZCI6MH0"
```

## Output

Returns JSON with an array of matched messages and a `cursor` field for pagination. Each message includes the sender, date, text, chat info, and message ID.

## See Also

- [tgs search global]({{< relref "/reference/commands/search-global" >}})
- [tgs search counters]({{< relref "/reference/commands/search-counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}})
