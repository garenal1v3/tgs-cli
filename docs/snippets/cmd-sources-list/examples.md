```bash
# List all dialogs
tgs sources list

# List only channels and supergroups
tgs sources list --type channel,supergroup

# List with full stats (slower — ~4 API calls per source)
tgs sources list --with-stats

# Include archived dialogs
tgs sources list --archived

# Paginate through a large account (first page)
tgs sources list --limit 50

# Fetch the next page using a cursor from the previous response
tgs sources list --limit 50 --cursor "eyJvIjo1MCwiZCI6MH0"
```
