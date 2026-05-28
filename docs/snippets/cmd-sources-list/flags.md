| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | | string[] | (all) | Filter by type: `channel`, `supergroup`, `group`, `user`, `bot` (repeatable, comma-separated) |
| `--with-stats` | | bool | `false` | Fetch expensive metrics (total/24h/first messages + full info) for every returned source |
| `--limit` | `-l` | int | `0` | Max records to return (1-500, 0=all) |
| `--cursor` | | string | | Pagination cursor from previous response |
| `--archived` | | bool | `false` | Include archived dialogs |
| `--max-wait` | | int | `60` | Max seconds to wait on FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Disable peer and stats caches |
| `--profile` | `-p` | string | | Account profile name |
