```bash
# Inspect a channel by username
tgs sources inspect @durov

# Inspect a group by numeric ID
tgs sources inspect -1009876543210

# Inspect Saved Messages (your own chat)
tgs sources inspect -

# Inspect a channel you are NOT subscribed to
tgs sources inspect @somechannel

# Skip expensive stats
tgs sources inspect @durov --no-stats

# Use a specific profile
tgs sources inspect @golang --profile work
```
