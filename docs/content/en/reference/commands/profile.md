---
title: tgs profile
weight: 20
---

# tgs profile

Manage account profiles.

## Synopsis

```
tgs profile <subcommand> [flags]
```

## Subcommands

### tgs profile list

List all profiles with their associated account information.

```
tgs profile list [flags]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `json` | Output format: `json`, `text` |

**Examples:**

```bash
tgs profile list
tgs profile list --output text
```

Text output marks the active profile with `*`:

```
* default (+1234567890, @myuser)
  work (+0987654321, @workuser)
```

JSON output includes an `active` boolean field per profile.

---

### tgs profile switch

Set the active profile for the current directory by writing a `.tgs.yaml` file.

```
tgs profile switch <name>
```

**Examples:**

```bash
tgs profile switch work
tgs profile switch default
```

Creates or overwrites `.tgs.yaml` in the current directory with:

```yaml
profile: <name>
```

All tgs commands run in this directory (and subdirectories) will use the specified profile.

---

### tgs profile delete

Delete a profile and its local session database.

```
tgs profile delete <name>
```

**Examples:**

```bash
tgs profile delete old-account
```

You cannot delete the currently active profile. Switch to a different profile first:

```bash
tgs profile switch default
tgs profile delete work
```

## See Also

- [tgs whoami](/reference/commands/whoami/)
- [Profiles guide](/guide/profiles/)
