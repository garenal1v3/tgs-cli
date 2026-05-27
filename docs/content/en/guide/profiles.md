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

```bash
cd ~/projects/work-project
tgs profile switch work
```

This writes a `.tgs.yaml` in the current directory:

```yaml
profile: work
```

tgs walks up the directory tree to find `.tgs.yaml`, so it applies in all subdirectories too. This is the recommended way to use different accounts per project.

## Managing Profiles

### List profiles

```bash
tgs profile list
```

Text output (`--output text`):

```
* default (+1234567890, @myuser)
  work (+0987654321, @workuser)
```

The `*` marks the currently active profile. JSON output is the default and includes an `active` field.

### Switch profile for current directory

```bash
tgs profile switch work
```

This creates or overwrites `.tgs.yaml` in the current directory.

### Delete a profile

```bash
tgs profile delete old-account
```

Deletes the profile directory and its session database. You cannot delete the currently active profile — switch to another profile first:

```bash
tgs profile switch default
tgs profile delete work
```

### Check current profile and account

```bash
tgs whoami
tgs whoami --output text
```

Text output:

```
Profile: work
User:    John Doe
Handle:  @johndoe
Phone:   +1234567890
ID:      123456789
```

`whoami` reads from local storage — it does not connect to Telegram.

## Environment Variable

Set `TGS_PROFILE` to override the profile for a single command or for the whole shell session:

```bash
# One-off override
TGS_PROFILE=work tgs whoami

# Override for the shell session
export TGS_PROFILE=work
```

The `--profile` flag always takes precedence over `TGS_PROFILE`.

## Practical Examples

### Use a work account for one command

```bash
tgs --profile work search "design review"
```

### Set up a project directory

```bash
mkdir ~/projects/client-x && cd ~/projects/client-x
tgs login --type code --profile client-x
tgs profile switch client-x
# All tgs commands in this directory now use client-x
```

### Check what profile is active

```bash
tgs whoami --output text
```

## Full Reference

- [tgs profile](/reference/commands/profile/) — list, switch, delete subcommands
- [tgs whoami](/reference/commands/whoami/) — show current account info
- [Environment Variables](/reference/environment/) — all supported env vars
