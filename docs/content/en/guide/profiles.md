---
title: Profiles
weight: 20
---

# Profiles

tgs supports multiple Telegram accounts through profiles — similar to how AWS CLI handles credentials. Each profile has its own independent session stored in a separate database file.

## How Profile Resolution Works

When you run any tgs command, the active profile is determined by this priority chain (highest wins):

1. `--profile` flag passed to the command
2. `TGS_PROFILE` environment variable
3. `.tgs.yaml` file in the current directory or any parent directory
4. `"default"` (fallback)

## Per-Directory Profiles

You can bind a profile to a directory by creating a `.tgs.yaml` file. Use `tgs profile switch` to do this automatically:

{{< snippet "profiles/switch-yaml.md" >}}

tgs walks up the directory tree to find `.tgs.yaml`, so it applies in all subdirectories too. This is the recommended way to use different accounts per project.

## Managing Profiles

### List profiles

{{< snippet "profiles/list.md" >}}

The `*` marks the currently active profile. JSON output is the default and includes an `active` field.

### Switch profile for current directory

{{< snippet "profiles/switch.md" >}}

This creates or overwrites `.tgs.yaml` in the current directory.

### Delete a profile

{{< snippet "profiles/delete.md" >}}

Deletes the profile directory and its session database. You cannot delete the currently active profile — switch to another profile first.

### Check current profile and account

{{< snippet "profiles/whoami.md" >}}

`whoami` reads from local storage — it does not connect to Telegram.

## Environment Variable

Set `TGS_PROFILE` to override the profile for a single command or for the whole shell session:

{{< snippet "profiles/env.md" >}}

The `--profile` flag always takes precedence over `TGS_PROFILE`.

## Practical Examples

{{< snippet "profiles/practical.md" >}}

## Full Reference

- [tgs profile]({{< relref "/reference/commands/account/profile" >}}) — list, switch, delete subcommands
- [tgs whoami]({{< relref "/reference/commands/account/whoami" >}}) — show current account info
- [Environment Variables]({{< relref "/reference/environment" >}}) — all supported env vars
