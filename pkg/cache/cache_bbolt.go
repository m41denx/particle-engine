package cache

import (
	"go.etcd.io/bbolt"
	"path/filepath"
)

type BboltCache struct {
	db *bbolt.DB
}

func NewBboltCache(homedir string) (*BboltCache, error) {
	bb, err := bbolt.Open(filepath.Join(homedir, "cache.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	bb.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket([]byte("cache"))
		return err
	})
	return &BboltCache{db: bb}, nil
}

func (c *BboltCache) Close() error {
	return c.db.Close()
}

func (c *BboltCache) Get(key []byte) ([]byte, error) {
	var val []byte
	err := c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		val = b.Get(key)
		return nil
	})
	return val, err
}

func (c *BboltCache) Set(key, val []byte) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		return b.Put(key, val)
	})
}

func (c *BboltCache) Delete(key []byte) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		return b.Delete(key)
	})
}
