```bash
# Inspect a channel by username
tgs sources inspect @durov

# Inspect a group by numeric ID — use the `id:` prefix (a bare "-1001…"
# would be parsed as a flag by the CLI; `id:` or a `--` separator avoids it)
tgs sources inspect id:-1009876543210
tgs sources inspect -- -1009876543210

# Inspect Saved Messages (your own chat)
tgs sources inspect -

# Inspect a channel you are NOT subscribed to
tgs sources inspect @somechannel

# Skip expensive stats
tgs sources inspect @durov --no-stats

# Use a specific profile
tgs sources inspect @golang --profile work
```
