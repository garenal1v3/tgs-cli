| Variable | Description |
|----------|-------------|
| `TGS_PROFILE` | Active profile name. Overrides `.tgs.yaml` but is overridden by `--profile` flag. |
| `TGS_API_ID` | Telegram API ID. Overrides the value compiled into the binary. |
| `TGS_API_HASH` | Telegram API Hash. Overrides the value compiled into the binary. |
| `TGS_CONFIG_DIR` | Config directory. Takes precedence over `XDG_CONFIG_HOME`. Default: `$XDG_CONFIG_HOME/tgs`, falling back to `~/.config/tgs`. |
| `TGS_DATA_DIR` | Data directory for session databases. Takes precedence over `XDG_DATA_HOME`. Default: `$XDG_DATA_HOME/tgs`, falling back to `~/.local/share/tgs`. |
| `XDG_CONFIG_HOME` | Base config directory (XDG spec). Used as `$XDG_CONFIG_HOME/tgs` when `TGS_CONFIG_DIR` is unset. |
| `XDG_DATA_HOME` | Base data directory (XDG spec). Used as `$XDG_DATA_HOME/tgs` when `TGS_DATA_DIR` is unset. |