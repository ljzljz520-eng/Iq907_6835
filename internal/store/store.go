package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.etcd.io/bbolt"

	"volunteertraining/internal/domain"
)

var (
	bucketUsers    = []byte("users")
	bucketVideos   = []byte("videos")
	bucketProgress = []byte("progress")
	bucketFeedback = []byte("feedback")
	bucketAudit    = []byte("audit")
)

type Store struct {
	db   *bbolt.DB
	path string
	mu   sync.RWMutex
}

func Open(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("%w: empty database path", domain.ErrInvalid)
	}
	db, err := bbolt.Open(filepath.Clean(path), 0o600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt database: %w", err)
	}
	s := &Store{db: db, path: path}
	if err := s.ensureBuckets(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) ensureBuckets() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{bucketUsers, bucketVideos, bucketProgress, bucketFeedback, bucketAudit} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("create bucket %q: %w", bucket, err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string { return s.path }

func (s *Store) view(fn func(*bbolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return fmt.Errorf("%w: store is closed", domain.ErrConflict)
	}
	return s.db.View(fn)
}

func (s *Store) update(fn func(*bbolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return fmt.Errorf("%w: store is closed", domain.ErrConflict)
	}
	return s.db.Update(fn)
}

func (s *Store) put(bucket []byte, key string, value any) error {
	encoded, err := marshal(value)
	if err != nil {
		return err
	}
	return s.update(func(tx *bbolt.Tx) error { return tx.Bucket(bucket).Put([]byte(key), encoded) })
}

func (s *Store) get(bucket []byte, key string, target any) error {
	return s.view(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucket).Get([]byte(key))
		if value == nil {
			return domain.ErrNotFound
		}
		return unmarshal(value, target)
	})
}

func (s *Store) list(bucket []byte, factory func() any, keyOf func(any) string) ([]any, error) {
	items := make([]any, 0)
	err := s.view(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			target := factory()
			if err := unmarshal(value, target); err != nil {
				return err
			}
			items = append(items, target)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return keyOf(items[i]) < keyOf(items[j]) })
	return items, nil
}

func stableID(prefix string, values ...string) string {
	h := sha256.New()
	for _, value := range values {
		_, _ = h.Write([]byte(value))
		_, _ = h.Write([]byte{0})
	}
	return prefix + "-" + hex.EncodeToString(h.Sum(nil)[:10])
}

func cloneBytes(value []byte) []byte { return bytes.Clone(value) }
