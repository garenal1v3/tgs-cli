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
| `--chat` | `-c` | string[] | | Chat to search (username, phone, ID; repeatable, comma-separated). Required unless `--folder` is set. |
| `--folder` | | string | `""` | Search all chats inside this folder (id or name); peers are additive to any `--chat` values |
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

Search across all chats in a folder:

```bash
tgs search messages "announcement" --folder Work
```

Paginate through results:

```bash
tgs search messages "bug" -c @dev -l 10 --cursor "eyJvIjo1MCwiZCI6MH0"
```

## Output

Returns JSON with an array of matched messages and a `cursor` field for pagination.

**JSON (default):**

```json
{
  "messages": [
    {
      "id": 520,
      "chat": {"id": 1006503122, "type": "channel", "title": "Pavel Durov"},
      "date": "2026-05-23T13:26:07Z",
      "text": "WhatsApp encryption is a giant fraud...",
      "media": {"type": "webpage"},
      "views": 1128733,
      "forwards": 10748,
      "replies": 540,
      "reactions": [
        {"emoji": "👍", "count": 8412},
        {"emoji": "🔥", "count": 2103},
        {"emoji": "❤", "count": 991}
      ]
    }
  ],
  "total": 285,
  "cursor": "eyJvIjo1MjAsImQiOjB9"
}
```

**Text (`--output text`):**

```
[2026-05-23 13:26:07] Pavel Durov: WhatsApp encryption is a giant fraud... [👁 1128733 ↻ 10748 💬 540 👍 8412 🔥 2103 ❤ 991]
```

### Message fields

| Field | Type | Notes |
|---|---|---|
| `id` | int | Message ID within the chat |
| `chat` | object | `{id, type, title?, username?}` — `type` is one of `channel`, `supergroup`, `group`, `private` |
| `from` | object | Sender info (omitted for channel posts, present for group messages and comments) |
| `date` | string | RFC3339 UTC timestamp |
| `text` | string | Plain text of the message |
| `media` | object | Type and metadata for attached media; omitted for text-only messages |
| `reply_to_msg_id` | int | Message this is a reply to; in discussion groups, points to the forwarded channel post |
| `topic_id` | int | Forum topic ID if applicable |
| `views`, `forwards`, `replies` | int | Engagement counters (channel posts) |
| `reactions` | array | List of `{emoji, count}` pairs; custom emoji appear as `custom:<doc_id>`, paid as `⭐` |
| `cursor` | string | Pagination cursor; absent when there are no more results |

## See Also

- [tgs search global]({{< relref "/reference/commands/search/global" >}})
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}})
