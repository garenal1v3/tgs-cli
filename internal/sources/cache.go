package sources

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var bucketSourceStats = []byte("source_stats")

// SourceStatsCache stores expensive per-source metrics in a BoltDB bucket. It
// is invalidated by last_message_id mismatch — there is no TTL.
type SourceStatsCache struct {
	db *bolt.DB
}

// cachedEntry is the on-disk JSON value.
type cachedEntry struct {
	Payload              *CachedPayload `json:"payload"`
	CachedAt             int64          `json:"cached_at"`
	LastMessageIDAtCache int            `json:"last_message_id_at_cache"`
}

// NewSourceStatsCache opens (or creates) a BoltDB at dbPath and ensures the
// source_stats bucket exists.
func NewSourceStatsCache(dbPath string) (*SourceStatsCache, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}
	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketSourceStats)
		return err
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create bucket: %w", err)
	}
	return &SourceStatsCache{db: db}, nil
}

// Put writes a payload for the given source, recording the last_message_id at
// the time of caching for later invalidation.
func (c *SourceStatsCache) Put(sourceID int64, lastMessageID int, payload *CachedPayload) error {
	entry := cachedEntry{
		Payload:              payload,
		CachedAt:             time.Now().Unix(),
		LastMessageIDAtCache: lastMessageID,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}
	return c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSourceStats).Put(keyFor(sourceID), data)
	})
}

// Get returns the cached payload only if currentLastMessageID matches the one
// stored at cache time. Otherwise it returns found=false.
func (c *SourceStatsCache) Get(sourceID int64, currentLastMessageID int) (*CachedPayload, bool, error) {
	var raw []byte
	err := c.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketSourceStats).Get(keyFor(sourceID))
		if v != nil {
			raw = make([]byte, len(v))
			copy(raw, v)
		}
		return nil
	})
	if err != nil {
		return nil, false, fmt.Errorf("read cache: %w", err)
	}
	if raw == nil {
		return nil, false, nil
	}
	var entry cachedEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, false, fmt.Errorf("unmarshal cache entry: %w", err)
	}
	if entry.LastMessageIDAtCache != currentLastMessageID {
		return nil, false, nil
	}
	return entry.Payload, true, nil
}

// Close closes the underlying BoltDB.
func (c *SourceStatsCache) Close() error {
	return c.db.Close()
}

// keyFor converts an int64 source ID to an 8-byte big-endian key.
func keyFor(sourceID int64) []byte {
	var k [8]byte
	binary.BigEndian.PutUint64(k[:], uint64(sourceID))
	return k[:]
}
