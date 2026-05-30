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
| `--folder` | | string | `""` | Filter to chats inside this folder (id or name), archived chats included; incompatible with `--cursor` |
| `--archived` | | bool | `false` | Include archived dialogs (ignored when `--folder` is set — a folder already includes its archived chats) |
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

List only chats inside the "Crypto" folder (archived members included):

```bash
tgs sources list --folder Crypto
```

Paginate through a large account:

```bash
# First page
tgs sources list --limit 50

# Next page using the cursor from the previous response
tgs sources list --limit 50 --cursor "eyJvIjo1MCwiZCI6MH0"
```

## Output

Returns a JSON object with a `sources` array, a `total` count, a `returned`
count, and a `cursor` for pagination.

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
  "returned": 2,
  "cursor": "eyJvIjo1MCwiZCI6MH0"
}
```

`total` is the unfiltered count (all dialogs in the folder(s) walked), while
`returned` is `len(sources)` after `--type` filtering and `--limit` trimming.
With a `--type` filter you'll typically see `returned < total`.

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
| `creation_date` | string | RFC3339 UTC timestamp of when a chat/channel was created; only present for channels/groups with `--with-stats` or via `inspect` |
| `invite_link` | string | Primary invite link; present when available |
| `first_name` | string | User/bot first name; omitted for channels/groups |
| `last_name` | string | User last name; omitted if unset |
| `phone` | string | E.164 phone digits (no `+`); omitted unless contact-visible |
| `verified` | bool | Official verified account; only present if `true` |
| `scam` | bool | Marked as scam by Telegram; only present if `true` |
| `fake` | bool | Marked as fake by Telegram; only present if `true` |
| `restricted` | bool | Restricted in some regions; only present if `true` |
| `restricted_reason` | string | Free-form restriction reason; omitted if absent |
| `deleted` | bool | Deleted user account; only present if `true` |
| `archived` | bool | Dialog is archived; only present if `true` |
| `pinned` | bool | Dialog is pinned; only present if `true` |
| `saved` | bool | This is Saved Messages; only present if `true` |
| `gigagroup` | bool | Broadcast group (gigagroup); only present if `true` |
| `has_topics` | bool | Supergroup with forum topics; only present if `true` |
| `unread_count` | int | Unread message count (always present, including `0`) |
| `last_message` | object | `{id, date}` of the most recent message |
| `stats` | object | Present only when `--with-stats` is set |
| `stats_error` | string | Short Telegram error code (e.g. `CHANNEL_PRIVATE`) when stats fetch failed; omitted on success |

The `cursor` field is **omitted from the JSON entirely** when there are no more results (it is not present with value `""`). Iterate with `jq -r '.cursor? // empty'` to terminate the loop cleanly.

### Performance note

By default, `tgs sources list` is fast — it reads dialogs from your account list. Adding `--with-stats` triggers roughly 4 additional API calls per source: `channels.getFullChannel` / `messages.getFullChat` / `users.getFullUser` for the full info, `messages.search` for the total count, plus paginated `messages.getHistory` calls for the 24-hour count and the first message. For accounts with hundreds of dialogs the full run can take several minutes. Stats for inactive sources (last message older than 7 days) are cached on disk and reused across runs.

### Stats accuracy

- `total_messages` is the server-reported message count for the dialog.
- `messages_24h` walks recent history (up to ~1000 messages) and counts those within the last 24 h; hyper-active peers are capped at that walking limit.
- `first_message` is best-effort — for broadcast channels Telegram's API may return no oldest message, in which case the field is omitted.

## See Also

- [tgs sources inspect]({{< relref "/reference/commands/sources/inspect" >}})
