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
