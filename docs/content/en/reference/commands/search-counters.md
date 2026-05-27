---
title: tgs search counters
weight: 12
---

# tgs search counters

Get message counts grouped by filter type (photo, video, document, etc.) for a specific chat.

## Usage

```
tgs search counters [flags]
```

This command takes no positional arguments.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--chat` | `-c` | string | | Chat to query (username, phone, or ID). **Required.** |
| `--topic` | | int | `0` | Forum topic ID |
| `--filters` | | string[] | | Filter types to count (comma-separated; default: all). See [filter values](/en/reference/commands/search-messages/#filter-values) |
| `--max-wait` | | int | `60` | Max seconds to wait on FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Disable peer resolution cache |
| `--profile` | `-p` | string | | Account profile name |

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env -> `.tgs.yaml` file -> `"default"`.

## Examples

Get all message type counts for a channel:

```bash
tgs search counters -c @golang
```

Get counts for specific filter types:

```bash
tgs search counters -c @mygroup --filters photo,video,document
```

## Output

Returns JSON with an array of objects, each containing a filter type name and the corresponding message count.

## See Also

- [tgs search messages]({{< relref "/reference/commands/search-messages" >}})
- [tgs search global]({{< relref "/reference/commands/search-global" >}})
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}})
