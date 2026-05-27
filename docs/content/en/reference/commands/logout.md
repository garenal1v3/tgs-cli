---
title: tgs logout
weight: 11
---

# tgs logout

Revoke the current session on Telegram servers and clear local session data.

## Synopsis

{{< snippet "cmd-logout/synopsis.md" >}}

## Flags

{{< snippet "cmd-logout/flags.md" >}}

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env → `.tgs.yaml` file → `"default"`.

## Examples

{{< snippet "cmd-logout/examples.md" >}}

## Notes

Logout performs two steps:

1. Calls the Telegram API to revoke the session token (the session becomes invalid on Telegram's side)
2. Clears the local session database for the profile

After logout, you must run `tgs login` again to use tgs with that profile.

## See Also

- [tgs login]({{< relref "/reference/commands/login" >}})
- [Authentication guide]({{< relref "/getting-started/authentication" >}})
