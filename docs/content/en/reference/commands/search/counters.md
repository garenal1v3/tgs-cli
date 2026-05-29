---
title: tgs search counters
weight: 30
---

# tgs search counters

Get message counts grouped by filter type (photo, video, document, etc.) for one or more chats.

## Usage

```
tgs search counters [flags]
```

This command takes no positional arguments.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--chat` | `-c` | string | | Chat to query (username, phone, or ID). Required unless `--folder` is set. |
| `--folder` | | string | `""` | Query all chats inside this folder (id or name); results include per-chat breakdowns plus aggregated totals |
| `--topic` | | int | `0` | Forum topic ID |
| `--filters` | | string[] | | Filter types to count (comma-separated; default: all). See [filter values](/en/reference/commands/search/messages/#filter-values) |
| `--max-wait` | | int | `60` | Max seconds to wait on FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Disable peer resolution cache |
| `--profile` | `-p` | string | | Account profile name |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env -> `.tgs.yaml` file -> `"default"`.

## Examples

Get all message type counts for a channel:

```bash
tgs search counters -c @golang
```

Get counts for specific filter types:

```bash
tgs search counters -c @mygroup --filters photo,video,document
```

Get counts for all chats in a folder:

```bash
tgs search counters --folder Work --filters photo,document
```

## Output

Returns JSON with an array of `{filter, count}` entries.

**JSON (default, single chat):**

```json
{
  "counters": [
    {"filter": "photo", "count": 98},
    {"filter": "video", "count": 45},
    {"filter": "photo-video", "count": 143},
    {"filter": "document", "count": 0},
    {"filter": "url", "count": 188},
    {"filter": "gif", "count": 10},
    {"filter": "voice", "count": 0},
    {"filter": "music", "count": 0},
    {"filter": "round-video", "count": 0},
    {"filter": "geo", "count": 0}
  ]
}
```

**JSON (with `--folder`, multi-chat):**

When `--folder` is used, the output includes a per-chat breakdown and aggregated totals:

```json
{
  "chats": [
    {
      "chat": {"id": -1001006503122, "type": "channel", "title": "Dev News"},
      "counters": [
        {"filter": "photo", "count": 30},
        {"filter": "document", "count": 12}
      ]
    },
    {
      "chat": {"id": -1001009876543, "type": "supergroup", "title": "Team"},
      "counters": [
        {"filter": "photo", "count": 68},
        {"filter": "document", "count": 41}
      ]
    }
  ],
  "totals": [
    {"filter": "photo", "count": 98},
    {"filter": "document", "count": 53}
  ]
}
```

**Text (`--output text`):**

```
photo        98
video        45
photo-video  143
document     0
url          188
gif          10
voice        0
music        0
round-video  0
geo          0
```

Telegram returns only the filter types it supports for the given chat — channels typically omit `contact`, `pinned`, `mention`, `phone-call`, and `chat-photo`.

## See Also

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}})
- [tgs search global]({{< relref "/reference/commands/search/global" >}})
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}})
