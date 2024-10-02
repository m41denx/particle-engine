package cache

type Cache interface {
	Get(key []byte) ([]byte, error)
	Set(key, val []byte) error
	Delete(key []byte) error
	Close() error
}
