```json
{
  "sources": [
    {
      "id": -1001234567890,
      "type": "channel",
      "title": "Durov's Channel",
      "username": "durov",
      "access": "public",
      "members_count": 1234567,
      "verified": true,
      "unread_count": 0,
      "last_message": {"id": 4321, "date": "2026-05-28T08:15:00Z"}
    },
    {
      "id": -1009876543210,
      "type": "supergroup",
      "title": "Go Programming",
      "username": "golang",
      "access": "public",
      "members_count": 78432,
      "has_topics": true,
      "unread_count": 5,
      "last_message": {"id": 120450, "date": "2026-05-28T07:42:11Z"}
    }
  ],
  "total": 287,
  "returned": 2,
  "cursor": "eyJvIjo1MCwiZCI6MH0"
}
```

With `--with-stats`:

```json
{
  "sources": [
    {
      "id": -1001234567890,
      "type": "channel",
      "title": "Durov's Channel",
      "username": "durov",
      "access": "public",
      "members_count": 1234567,
      "verified": true,
      "unread_count": 0,
      "last_message": {"id": 4321, "date": "2026-05-28T08:15:00Z"},
      "creation_date": "2015-08-26T10:00:00Z",
      "stats": {
        "total_messages": 12345,
        "messages_24h": 3,
        "first_message": {"id": 1, "date": "2015-08-26T10:00:00Z"}
      }
    }
  ],
  "total": 287,
  "returned": 1
}
```

> The `cursor` key is omitted from the JSON when the listing is complete; it appears only when more pages are available.
