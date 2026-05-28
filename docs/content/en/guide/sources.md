---
title: Sources
weight: 40
---

# Sources

`tgs sources` lets you enumerate and inspect the Telegram dialogs in your account — channels, supergroups, groups, users, and bots. All output is JSON, ready to pipe into `jq` or feed to an AI agent.

> **Piping to `jq`:** `tgs` writes diagnostic logs (`[tgs] retry: …`, `[tgs] FLOOD_WAIT: …`) to **stderr** and JSON to **stdout**. When piping into `jq`, drop stderr to keep the parser happy:
>
> ```bash
> tgs sources list --with-stats 2>/dev/null | jq '.sources[].title'
> ```
>
> Avoid `2>&1 | jq …` — that mixes log lines into stdout and breaks JSON parsing.

## Listing all your sources

The simplest invocation returns every dialog in your account:

```bash
tgs sources list
```

The response is a JSON object with a `sources` array, a `total` count, and an optional `cursor`:

```json
{
  "sources": [
    {
      "id": -1001234567890,
      "type": "channel",
      "title": "Durov's Channel",
      "username": "durov",
      "access": "public",
      "members_count": 1234567,
      "verified": true,
      "unread_count": 0,
      "last_message": {"id": 4321, "date": "2026-05-28T08:15:00Z"}
    }
  ],
  "total": 287
}
```

> **`total`** is the dialog count as known to your account — it reflects all dialogs **before** any `--type` filtering. The CLI guarantees `total >= len(sources)` (Telegram's own counter is sometimes stale; we bump it up to match what we actually returned). Use `cursor` for pagination — don't divide by `total`.

## Filtering by type

Use `--type` to narrow results to specific dialog types. You can pass a single type or a comma-separated list:

```bash
# Only channels
tgs sources list --type channel

# Channels and supergroups
tgs sources list --type channel,supergroup

# Only bots
tgs sources list --type bot
```

Supported types: `channel`, `supergroup`, `group`, `user`, `bot`.

## Getting full metrics

By default, `tgs sources list` is fast and cheap — it reads your dialog list without extra API calls. Add `--with-stats` to also fetch message statistics for every source:

```bash
tgs sources list --with-stats
```

Each source in the response will then include a `stats` object:

```json
{
  "stats": {
    "total_messages": 12345,
    "messages_24h": 3,
    "first_message": {"id": 1, "date": "2015-08-26T10:00:00Z"}
  }
}
```

`first_message` is best-effort: for broadcast channels Telegram's MTProto API may return no oldest message at all (the field is then omitted). `total_messages` and `messages_24h` are always present.

**Performance note:** `--with-stats` triggers roughly 4 API calls per source, and `messages_24h` paginates message history backwards (up to ~1000 messages) until it crosses the 24-hour cutoff — extremely active sources are capped at that value. For large accounts the full run can take several minutes. Stats for inactive sources (last message older than 7 days) are cached on disk and reused between runs.

## Inspecting a single source

`tgs sources inspect` gives you the full picture for one source. It works for channels you have joined and for public channels you haven't:

```bash
# By username
tgs sources inspect @durov

# By numeric Telegram ID — use the `id:` prefix (a bare "-1001234…" is eaten
# by the CLI's flag parser; use `id:` or insert `--` to disambiguate).
tgs sources inspect id:-1001234567890

# `--` works too, but everything after it is treated as positional, so any
# command flags must come BEFORE the separator:
tgs sources inspect --no-stats -- -1001234567890

# Your Saved Messages
tgs sources inspect -
```

The response includes extra fields not available in `list`: `subscribed`, `description`, `creation_date`, and `invite_link`. For broadcast channels, `creation_date` reflects when the channel was created; for users, it is omitted (Telegram does not expose user registration date).

> **Numeric ID limitation:** Numeric IDs only work if the peer was previously resolved by `@username` or `+phone` (which populates the local peer cache with the required access hash). Plain `tgs sources list` does NOT seed the cache. If a numeric-ID inspect fails with "peer is not in the local cache", run one `inspect @<username>` (or `+<phone>`) first, then retry with `id:<n>`.

### Inspecting a channel you're not subscribed to

As long as you know the `@username`, you can inspect any public channel without joining it:

```bash
tgs sources inspect @somebigtechchannel
```

`subscribed` will be `false` in the response, and `stats` will be omitted. Other display fields — `title`, `username`, `access`, `members_count`, `description`, `verified` — are populated from the public channel info.

To skip fetching stats (faster):

```bash
tgs sources inspect @durov --no-stats
```

## Working with cursors for large accounts

If you have hundreds of dialogs, use `--limit` and `--cursor` for paginated access:

```bash
# First page
tgs sources list --limit 50
```

The response contains a `cursor` string. Pass it back to fetch the next page:

```bash
tgs sources list --limit 50 --cursor "eyJvIjo1MCwiZCI6MH0"
```

When there are no more pages, the `cursor` key is **omitted from the JSON entirely** (it is not present with value `""`). The pagination loop above handles this correctly via `jq -r '.cursor? // empty'`.

A typical pagination loop in a shell script:

```bash
cursor=""
while true; do
  if [ -z "$cursor" ]; then
    result=$(tgs sources list --limit 50)
  else
    result=$(tgs sources list --limit 50 --cursor "$cursor")
  fi

  echo "$result" | jq '.sources[]'

  cursor=$(echo "$result" | jq -r '.cursor? // empty')
  [ -z "$cursor" ] && break
done
```

## Archived dialogs

Archived dialogs are hidden from the default list. Pass `--archived` to include them:

```bash
tgs sources list --archived
```

## Full Reference

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}}) — all flags and output fields
- [tgs sources inspect]({{< relref "/reference/commands/sources/inspect" >}}) — inspect a single source
