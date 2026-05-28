---
title: Sources
weight: 40
---

# Sources

`tgs sources` lets you enumerate and inspect the Telegram dialogs in your account — channels, supergroups, groups, users, and bots. All output is JSON, ready to pipe into `jq` or feed to an AI agent.

## Listing all your sources

The simplest invocation returns every dialog in your account:

```bash
tgs sources list
```

The response is a JSON object with a `sources` array and a `total` count:

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
  "total": 287,
  "cursor": ""
}
```

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

**Performance note:** `--with-stats` triggers approximately 4 API calls per source. For large accounts this can take several minutes. Stats for inactive sources (last message older than 7 days) are cached on disk — subsequent runs reuse the cache until the source gets a new message.

## Inspecting a single source

`tgs sources inspect` gives you the full picture for one source. It works for channels you have joined and for public channels you haven't:

```bash
# By username
tgs sources inspect @durov

# By numeric Telegram ID
tgs sources inspect -1001234567890

# Your Saved Messages
tgs sources inspect -
```

The response includes extra fields not available in `list`: `subscribed`, `description`, `creation_date`, and `invite_link`.

### Inspecting a channel you're not subscribed to

As long as you know the `@username`, you can inspect any public channel without joining it:

```bash
tgs sources inspect @somebigtechchannel
```

`subscribed` will be `false` in the response, and `stats` will be `null`.

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

When the `cursor` field in the response is an empty string, you've reached the last page.

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
