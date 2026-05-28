---
title: tgs search global
weight: 20
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
| `--filter` | | string | | Message type filter (see [filter values](/en/reference/commands/search/messages/#filter-values)) |
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

Same shape as [tgs search messages]({{< relref "/reference/commands/search/messages" >}}#output) — a `messages` array, `total`, and optional `cursor`. Because results span many chats, each message's `chat` field tells you which chat it came from.

**JSON (default):**

```json
{
  "messages": [
    {
      "id": 188791,
      "chat": {"id": 1754252633, "type": "channel", "title": "News", "username": "newschannel"},
      "date": "2026-05-27T10:51:40Z",
      "text": "...",
      "views": 360330,
      "forwards": 283
    },
    {
      "id": 67578,
      "chat": {"id": 1069896405, "type": "supergroup", "title": "Tech Talk", "username": "techtalk"},
      "from": {"id": 1978176, "first_name": "Ilya", "username": "valkin"},
      "date": "2026-05-27T10:14:00Z",
      "text": "..."
    }
  ],
  "total": 2066,
  "cursor": "eyJvIjo0NzgxNSwiZCI6MCwiciI6MTc3OTg3MDc1Mn0"
}
```

Note: the cursor for global search includes an additional `r` (rate) field — pass it back verbatim with `--cursor`.

## See Also

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}})
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}})
