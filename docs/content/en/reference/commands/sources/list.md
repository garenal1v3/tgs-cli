---
title: tgs sources list
weight: 10
---

# tgs sources list

Enumerate all Telegram dialogs in your account — channels, supergroups, groups, users, and bots.

## Usage

```
tgs sources list [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | | string[] | (all) | Filter by type: `channel`, `supergroup`, `group`, `user`, `bot` (repeatable, comma-separated) |
| `--with-stats` | | bool | `false` | Fetch expensive metrics (total/24h/first messages + full info) for every returned source |
| `--limit` | `-l` | int | `0` | Max records to return (1-500, 0=all) |
| `--cursor` | | string | | Pagination cursor from previous response |
| `--archived` | | bool | `false` | Include archived dialogs |
| `--max-wait` | | int | `60` | Max seconds to wait on FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Disable peer and stats caches |
| `--profile` | `-p` | string | | Account profile name |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env -> `.tgs.yaml` file -> `"default"`.

## Examples

List all dialogs:

```bash
tgs sources list
```

Filter to channels and supergroups only:

```bash
tgs sources list --type channel,supergroup
```

Fetch stats for every source (slower — ~4 API calls per source):

```bash
tgs sources list --with-stats
```

Include archived dialogs:

```bash
tgs sources list --archived
```

Paginate through a large account:

```bash
# First page
tgs sources list --limit 50

# Next page using the cursor from the previous response
tgs sources list --limit 50 --cursor "eyJvIjo1MCwiZCI6MH0"
```

## Output

Returns a JSON object with a `sources` array, a `total` count, and a `cursor` for pagination.

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
    },
    {
      "id": -1009876543210,
      "type": "supergroup",
      "title": "Go Programming",
      "username": "golang",
      "access": "public",
      "members_count": 78432,
      "has_topics": true,
      "unread_count": 5,
      "last_message": {"id": 120450, "date": "2026-05-28T07:42:11Z"}
    }
  ],
  "total": 287,
  "cursor": "eyJvIjo1MCwiZCI6MH0"
}
```

When `--with-stats` is set, each source also includes a `stats` object:

```json
{
  "stats": {
    "total_messages": 12345,
    "messages_24h": 3,
    "first_message": {"id": 1, "date": "2015-08-26T10:00:00Z"}
  }
}
```

### Source fields

| Field | Type | Notes |
|---|---|---|
| `id` | int | Telegram peer ID (negative for channels/groups) |
| `type` | string | One of `channel`, `supergroup`, `group`, `user`, `bot` |
| `title` | string | Display name |
| `username` | string | Public username without `@`; omitted if not set |
| `access` | string | `public` or `private`; omitted for users |
| `members_count` | int | Member/subscriber count; omitted if unavailable |
| `description` | string | Bio or about text; only present with full info |
| `has_comments` | bool | Channel has a linked discussion group; only present if `true` |
| `linked_chat_id` | int | ID of the linked discussion group; only present if applicable |
| `verified` | bool | Official verified account; only present if `true` |
| `scam` | bool | Marked as scam by Telegram; only present if `true` |
| `fake` | bool | Marked as fake by Telegram; only present if `true` |
| `restricted` | bool | Restricted in some regions; only present if `true` |
| `archived` | bool | Dialog is archived; only present if `true` |
| `pinned` | bool | Dialog is pinned; only present if `true` |
| `saved` | bool | This is Saved Messages; only present if `true` |
| `gigagroup` | bool | Broadcast group (gigagroup); only present if `true` |
| `has_topics` | bool | Supergroup with forum topics; only present if `true` |
| `unread_count` | int | Unread message count |
| `last_message` | object | `{id, date}` of the most recent message |
| `stats` | object | Present only when `--with-stats` is set |

The `cursor` field in the response is empty when there are no more results.

### Performance note

By default, `tgs sources list` is fast — it reads dialogs from your account list. Adding `--with-stats` triggers approximately 4 additional API calls per source to fetch full info and message statistics. For accounts with hundreds of dialogs, this can take several minutes. Stats for inactive sources (last message older than 7 days) are cached on disk and reused across runs.

## See Also

- [tgs sources inspect]({{< relref "/reference/commands/sources/inspect" >}})
