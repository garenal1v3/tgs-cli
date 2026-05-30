---
title: tgs profile
weight: 10
---

# tgs profile

Manage account profiles.

## Synopsis

{{< snippet "cmd-profile/synopsis.md" >}}

## Subcommands

### tgs profile list

List all profiles with their associated account information.

{{< snippet "cmd-profile/list.md" >}}

**JSON (default):**

```json
[
  {"name":"default","phone":"+79001234567","username":"alice","id":261054642,"active":true},
  {"name":"work","phone":"+79001230000","username":"alice_work","id":342178890,"active":false}
]
```

**Text (`--output text`):**

```
* default (@alice)
  work (@alice_work)
```

The active profile is marked with `*` in text mode, and `"active": true` in JSON. If neither username nor phone is cached, the profile name alone is shown.

---

### tgs profile switch

Set the active profile for the current directory by writing a `.tgs.yaml` file.

{{< snippet "cmd-profile/switch.md" >}}

**JSON (default):**

```json
{"profile":"work","status":"switched","dir":"/Users/alice/projects"}
```

**Text (`--output text`):**

```
Switched to profile "work" (wrote .tgs.yaml in /Users/alice/projects)
```

All tgs commands run in this directory (and subdirectories) will use the specified profile.

---

### tgs profile delete

Delete a profile and its local session database.

{{< snippet "cmd-profile/delete.md" >}}

**JSON (default):**

```json
{"profile":"old","status":"deleted"}
```

**Text (`--output text`):**

```
Deleted profile "old"
```

You cannot delete the currently active profile. Switch to a different profile first.

## See Also

- [tgs whoami]({{< relref "/reference/commands/account/whoami" >}})
- [Profiles guide]({{< relref "/guide/profiles" >}})
