package cache

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"

	"github.com/jf-tech/omniparser/mathutil"
)

// LoadingCache is a key/value cache with a user specified loading function and an optional capacity.
// If a key isn't present in the cache, load function will be called to create a value for the key and
// the value will be stored into the cache as well as returned to the caller. If no capacity is specified,
// then a reasonable default cache capacity is used (to prevent accidental and unintentional unlimited
// cache growth). If 0 capacity is specified, then cache capacity is removed - be careful of unlimited
// cache growth. If >0 capacity is specified and reached, cache entry eviction takes place in LRU fashion.
// The cache is thread-safe.
type LoadingCache struct {
	capacity int
	cache    *lru.Cache
}

const (
	defaultCapacity = 65536
)

// NewLoadingCache creates a new LoadingCache.
func NewLoadingCache(capacity ...int) *LoadingCache {
	capv := defaultCapacity
	if len(capacity) > 0 {
		if capacity[0] < 0 {
			panic(fmt.Errorf("capacity must be >= 0, instead got: %d", capacity[0]))
		}
		capv = capacity[0]
	}
	if capv == 0 {
		capv = mathutil.MaxIntValue - 1
	}
	cache, _ := lru.New(capv)
	return &LoadingCache{capacity: capv, cache: cache}
}

// LoadFunc is the type of the loading function.
type LoadFunc func(key interface{}) (interface{}, error)

// Get tries to fetch the value for a key from the cache; if not found, it  will call
// the load function to create the value for the key, store it into the cache and return
// the value.
func (c *LoadingCache) Get(key interface{}, load LoadFunc) (interface{}, error) {
	if v, found := c.cache.Get(key); found {
		return v, nil
	}
	v, err := load(key)
	if err != nil {
		return nil, err
	}
	c.cache.Add(key, v)
	return v, nil
}

// DumpForTest returns all the entries in the cache. Not thread-safe and should
// really only be used in tests as the function name suggests.
func (c *LoadingCache) DumpForTest() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})
	keys := c.cache.Keys()
	for _, k := range keys {
		if v, found := c.cache.Get(k); found {
			m[k] = v
		}
	}
	return m
}
