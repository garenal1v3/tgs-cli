---
title: tgs sources folders
weight: 30
---

# tgs sources folders

Enumerates the user's Telegram **folders** (dialog filters from the UI) and
resolves each one's contents the way the Telegram client does — applying
pinned/include/exclude peers and auto-flags (contacts, groups, etc.) to the
full dialog list.

The default "All chats" folder is excluded; only user-defined folders and
shared chatlists appear in the output.

## Usage

{{< snippet "cmd-sources-folders/synopsis.md" >}}

## Flags

{{< snippet "cmd-sources-folders/flags.md" >}}

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env -> `.tgs.yaml` file -> `"default"`.

## Examples

{{< snippet "cmd-sources-folders/examples.md" >}}

## Output

{{< snippet "cmd-sources-folders/output.md" >}}

### Folder fields

| Field | Type | Notes |
|---|---|---|
| `id` | int | Telegram dialog filter ID |
| `kind` | string | `custom` for user-defined folders; `chatlist` for shared chatlists |
| `title` | string | Display name of the folder |
| `emoticon` | string | Emoji set for the folder; omitted if not set |
| `has_my_invites` | bool | Chatlist has invite links created by you; only present if `true` |
| `chats_count` | int | Number of chats in `chats` array |
| `chats` | array | Resolved chat entries (same shape as `sources list` sources) |
| `total` | int | Total number of folders returned |

Chat entries inside `chats` follow the same field shape as
[tgs sources list]({{< relref "/reference/commands/sources/list" >}}) sources.

## Notes

- For folders that include archived chats explicitly (e.g. via `include_peers`)
  but you did NOT pass `--archived`, those archived chats will not appear —
  the command did not walk the archive folder. Pass `--archived` for the full
  picture.
- Folder contents are not paginated. The full list of folders is small
  (Telegram caps at ~30) and each folder's chats are returned in one response.

## See Also

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}})
- [tgs sources inspect]({{< relref "/reference/commands/sources/inspect" >}})
