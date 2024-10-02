package cache

type FakeCache struct {
}

func NewFakeCache() *FakeCache {
	return &FakeCache{}
}

func (c *FakeCache) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (c *FakeCache) Set(key, val []byte) error {
	return nil
}

func (c *FakeCache) Delete(key []byte) error {
	return nil
}

func (c *FakeCache) Close() error {
	return nil
}
