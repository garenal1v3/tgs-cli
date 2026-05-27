package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketSession = []byte("session")
	bucketMeta    = []byte("meta")
	keySession    = []byte("data")
)

type Session struct {
	db *bolt.DB
}

func NewSession(dbPath string) (*Session, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketSession, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create buckets: %w", err)
	}

	return &Session{db: db}, nil
}

func (s *Session) LoadSession(_ context.Context) ([]byte, error) {
	var data []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSession)
		v := b.Get(keySession)
		if v != nil {
			data = make([]byte, len(v))
			copy(data, v)
		}
		return nil
	})
	return data, err
}

func (s *Session) StoreSession(_ context.Context, data []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSession).Put(keySession, data)
	})
}

func (s *Session) StoreMeta(key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMeta).Put([]byte(key), value)
	})
}

func (s *Session) LoadMeta(key string) ([]byte, error) {
	var data []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketMeta).Get([]byte(key))
		if v != nil {
			data = make([]byte, len(v))
			copy(data, v)
		}
		return nil
	})
	return data, err
}

func (s *Session) LoadAllMeta() (map[string]string, error) {
	result := make(map[string]string)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMeta).ForEach(func(k, v []byte) error {
			result[string(k)] = string(v)
			return nil
		})
	})
	return result, err
}

func (s *Session) Close() error {
	return s.db.Close()
}
