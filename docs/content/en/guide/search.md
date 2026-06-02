---
title: Search
weight: 30
---

# Search

tgs provides four search subcommands that map directly to Telegram's MTProto search methods. All output is JSON, making it straightforward to pipe into `jq`, feed to an AI agent, or process programmatically.

## Searching in a Chat

The most common operation: search for messages within a specific chat.

```bash
tgs search messages "docker compose" -c @devops_notes
```

The query argument is required. The `-c` / `--chat` flag specifies which chat to search. See [Chat Resolution](#chat-resolution) below for all supported formats.

## Multiple Chats

Pass multiple chats by repeating the flag or using comma-separated values:

```bash
# Repeated flag
tgs search messages "release" -c @frontend -c @backend

# Comma-separated
tgs search messages "release" -c @frontend,@backend
```

tgs runs the search against each chat sequentially and merges the results into a single JSON response.

## Filtering by Author

Use `--from` / `-f` to show only messages from a specific sender:

```bash
tgs search messages "bug" -c @dev_chat --from @alice
```

The `--from` value accepts the same formats as `--chat` (username, numeric ID, phone number, or `t.me` link).

## Filtering by Content Type

The `--filter` flag restricts results to a specific message type. There are 15 available filters:

| Filter | Description |
|---|---|
| `photo` | Photos |
| `video` | Videos |
| `photo-video` | Photos and videos combined |
| `document` | Files and documents |
| `url` | Messages containing URLs |
| `gif` | GIF animations |
| `voice` | Voice messages |
| `music` | Audio files / music |
| `round-video` | Video messages (circles) |
| `geo` | Locations and live locations |
| `contact` | Shared contacts |
| `pinned` | Pinned messages |
| `mention` | Messages mentioning you |
| `phone-call` | Phone calls |
| `chat-photo` | Chat photo changes |

Example -- find all documents in a channel:

```bash
tgs search messages "" -c @project_files --filter document
```

An empty query (`""`) with a filter returns all messages of that type.

## Date Filtering

Narrow results to a date range with `--after` and `--before`. Both accept two formats:

- **Date string**: `YYYY-MM-DD` (interpreted as midnight UTC)
- **Unix timestamp**: raw seconds since epoch

```bash
# Messages from January 2025
tgs search messages "deploy" -c @ops --after 2025-01-01 --before 2025-02-01

# Using unix timestamps
tgs search messages "deploy" -c @ops --after 1704067200 --before 1706745600
```

Both flags are optional and can be used independently.

## Forum Topics

For supergroups with forum mode enabled, use `--topic` to search within a specific topic by its ID:

```bash
tgs search messages "error" -c @dev_forum --topic 42
```

Without `--topic`, the search covers all topics in the forum.

## Searching Channel Comments

Telegram channels can have a linked discussion group where comments under each post live. Use `--include-comments` to automatically search both the channel and its discussion group with a single command -- no need to look up the group's username:

```bash
tgs search messages "release" -c @somechannel --include-comments
```

Each channel in `--chat` is checked for a linked discussion group; if it has one, that group is added to the search transparently. Channels without comments are unaffected.

Results merge posts and comments sorted by date. The `reply_to_msg_id` field on each comment points to the message it replies to, letting you reconstruct the discussion thread.

## Global Search

Search across all your chats at once with `tgs search global`:

```bash
tgs search global "meeting notes"
```

### Narrowing by Chat Type

Restrict global search to specific chat types:

```bash
# Only channels
tgs search global "announcement" --channels-only

# Only groups
tgs search global "discussion" --groups-only

# Only private chats
tgs search global "hey" --users-only
```

These flags are mutually exclusive -- passing more than one at once is rejected with an error.

### Searching in the archive

Pass `--archived` to restrict global search to chats in Telegram's built-in archive folder:

```bash
tgs search global "old thread" --archived
```

Global search also supports `--filter`, `--after`, `--before`, `--limit`, and `--cursor`, same as `search messages`.

## Content Counters

Get a breakdown of message counts by type for a chat without fetching the messages themselves:

```bash
tgs search counters -c @mychannel
```

This returns counts for all 15 filter types. To query only specific types:

```bash
tgs search counters -c @mychannel --filters photo,video,document
```

Counters also support `--topic` for forum groups.

## Calendar View

Get search results grouped by date -- useful for understanding when specific content was posted:

```bash
tgs search calendar -c @mychannel --filter photo
```

Both `--chat` and `--filter` are required for calendar queries. The response contains a list of dates with the corresponding message count and a message ID range (`min_msg_id`, `max_msg_id`) for each date.

## Filtering by folder

All search subcommands accept `--folder <id|name>`, which scopes the search to the chats inside a user-defined Telegram folder. `--folder` accepts a numeric folder ID or a case-insensitive folder name:

```bash
# Search messages across every chat in the "Work" folder
tgs search messages "release notes" --folder Work

# Count media in chats inside the "Crypto" folder
tgs search counters --folder Crypto

# Calendar view, scoped to a folder
tgs search calendar --folder Crypto --filter photo
```

Under the hood, `tgs` resolves the folder to its member chats and fans the request out across all of them, merging results just like a multi-chat search. The resolved chat list always includes the folder's **archived** chats — a folder is a view that can span the archive — so `--archived` is not needed (and is ignored) when `--folder` is set.

> **Note for `tgs search global`:** the old `--folder` flag on `search global` (which selected Telegram's main or archive folder by integer ID) has been replaced by `--archived`. The `--folder` flag now uniformly refers to user-defined folders across all subcommands.

To browse your folders and find IDs or exact names, use `tgs sources folders`.

## Cursor Pagination

tgs uses cursor-based pagination rather than offset-based. This matches Telegram's native approach and avoids skipped or duplicated results when new messages arrive during pagination.

Each search response includes a `cursor` field (empty string if there are no more results). Pass it back with `--cursor` to fetch the next page:

```bash
# First page (default: 50 results)
tgs search messages "update" -c @news --limit 10

# Next page using the cursor from the previous response
tgs search messages "update" -c @news --limit 10 --cursor "eyJvIjo1MCwiZCI6MH0"
```

The `--limit` flag controls page size (1-100, default 50).

A typical pagination loop in a script:

```bash
cursor=""
while true; do
  if [ -z "$cursor" ]; then
    result=$(tgs search messages "query" -c @chat --limit 20)
  else
    result=$(tgs search messages "query" -c @chat --limit 20 --cursor "$cursor")
  fi

  echo "$result" | jq '.messages[]'

  cursor=$(echo "$result" | jq -r '.cursor? // empty')
  [ -z "$cursor" ] && break
done
```

## Chat Resolution

tgs accepts multiple formats for identifying chats, users, and groups:

| Format | Example | Description |
|---|---|---|
| `@username` | `@durov` | Username with @ prefix |
| `username` | `durov` | Username without @ prefix |
| Numeric ID | `123456789` | Positive user/chat ID |
| Negative ID | `-1001234567890` | Supergroup/channel ID (Telegram's native format) |
| `t.me` link | `t.me/durov` | Telegram link (with or without `https://`) |
| Phone number | `+79001234567` | International phone format (7+ digits) |

All of these work in `--chat`, `--from`, and any other flag that accepts a peer reference.

> **Note:** Invite links (`t.me/+hash`) are not supported for search.

tgs caches resolved peers locally to avoid redundant API calls. Use `--no-cache` to bypass the cache if you suspect stale data. (`search global` resolves peers server-side and has no `--no-cache` flag.)

## Rate Limiting

Telegram enforces rate limits via `FLOOD_WAIT` errors that specify how many seconds you must wait before retrying. tgs handles this automatically:

1. When a `FLOOD_WAIT` is received, tgs prints a warning to stderr and sleeps for the required duration.
2. After the wait, it retries the request.
3. If the required wait exceeds `--max-wait` (default: 60 seconds), tgs returns the error immediately instead of blocking.

```bash
# Allow up to 120 seconds of flood wait
tgs search messages "query" -c @bigchannel --max-wait 120

# Fail fast -- don't wait more than 5 seconds
tgs search messages "query" -c @bigchannel --max-wait 5
```

For transient network errors, tgs retries up to 3 times with exponential backoff and jitter. Permanent errors (invalid peer, insufficient permissions, etc.) are never retried.

## Full Reference

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}}) -- search within specific chats
- [tgs search global]({{< relref "/reference/commands/search/global" >}}) -- search across all chats
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}}) -- message count breakdown by type
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}}) -- search results grouped by date
