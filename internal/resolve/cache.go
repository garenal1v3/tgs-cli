package resolve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var bucketPeers = []byte("peers")

// CacheEntry represents a resolved peer stored in the cache.
type CacheEntry struct {
	PeerType   string `json:"peer_type"`   // "user", "channel", "chat"
	ID         int64  `json:"id"`
	AccessHash int64  `json:"access_hash"`
	ResolvedAt int64  `json:"resolved_at"` // unix timestamp
}

// PeerCache is a BoltDB-backed cache for resolved Telegram peers.
type PeerCache struct {
	db  *bolt.DB
	TTL time.Duration
}

// NewPeerCache opens (or creates) a BoltDB at dbPath with a peers bucket and a default TTL of 24h.
func NewPeerCache(dbPath string) (*PeerCache, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketPeers)
		return err
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create peers bucket: %w", err)
	}

	return &PeerCache{db: db, TTL: 24 * time.Hour}, nil
}

// Store writes a CacheEntry under the given key.
func (c *PeerCache) Store(key string, entry CacheEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPeers).Put([]byte(key), data)
	})
}

// Load retrieves a CacheEntry by key. It returns found=false if the key does not exist
// or if the entry has expired according to the cache TTL.
func (c *PeerCache) Load(key string) (CacheEntry, bool, error) {
	var entry CacheEntry

	var raw []byte
	err := c.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketPeers).Get([]byte(key))
		if v != nil {
			raw = make([]byte, len(v))
			copy(raw, v)
		}
		return nil
	})
	if err != nil {
		return entry, false, fmt.Errorf("read cache: %w", err)
	}

	if raw == nil {
		return entry, false, nil
	}

	if err := json.Unmarshal(raw, &entry); err != nil {
		return entry, false, fmt.Errorf("unmarshal cache entry: %w", err)
	}

	// Check TTL expiration.
	if time.Since(time.Unix(entry.ResolvedAt, 0)) > c.TTL {
		return entry, false, nil
	}

	return entry, true, nil
}

// Close closes the underlying BoltDB.
func (c *PeerCache) Close() error {
	return c.db.Close()
}
