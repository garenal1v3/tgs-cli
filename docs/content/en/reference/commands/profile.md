---
title: tgs profile
weight: 20
---

# tgs profile

Manage account profiles.

## Synopsis

{{< snippet "cmd-profile/synopsis.md" >}}

## Subcommands

### tgs profile list

List all profiles with their associated account information.

{{< snippet "cmd-profile/list.md" >}}

Text output marks the active profile with `*`. JSON output includes an `active` boolean field per profile.

---

### tgs profile switch

Set the active profile for the current directory by writing a `.tgs.yaml` file.

{{< snippet "cmd-profile/switch.md" >}}

All tgs commands run in this directory (and subdirectories) will use the specified profile.

---

### tgs profile delete

Delete a profile and its local session database.

{{< snippet "cmd-profile/delete.md" >}}

You cannot delete the currently active profile. Switch to a different profile first.

## See Also

- [tgs whoami]({{< relref "/reference/commands/whoami" >}})
- [Profiles guide]({{< relref "/guide/profiles" >}})
