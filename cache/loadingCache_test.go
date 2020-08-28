package cache

import (
	"errors"
	"math/rand"
	"strconv"
	"testing"

	"github.com/antchfx/xpath"
	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/omniparser/mathutil"
)

func TestNewLoadingCache(t *testing.T) {
	for _, test := range []struct {
		name             string
		capacity         []int
		panicErr         string
		expectedCapacity int
	}{
		{
			name:             "no capacity specified, default used",
			capacity:         nil,
			panicErr:         "",
			expectedCapacity: defaultCapacity,
		},
		{
			name:             "0 capacity specified, no limit",
			capacity:         []int{0, -5}, // the second invalid value -5 is ignored.
			panicErr:         "",
			expectedCapacity: mathutil.MaxIntValue - 1,
		},
		{
			name:             "> 0 capacity specified",
			capacity:         []int{5},
			panicErr:         "",
			expectedCapacity: 5,
		},
		{
			name:             "invalid capacity specified",
			capacity:         []int{-1},
			panicErr:         "capacity must be >= 0, instead got: -1",
			expectedCapacity: 0,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.panicErr == "" {
				c := NewLoadingCache(test.capacity...)
				assert.NotNil(t, c)
				assert.Equal(t, test.expectedCapacity, c.capacity)
			} else {
				assert.PanicsWithError(t, test.panicErr, func() {
					NewLoadingCache(test.capacity...)
				})
			}
		})
	}
}

func TestLoadingCache_Get(t *testing.T) {
	for _, test := range []struct {
		name          string
		cache         *LoadingCache
		key           string
		load          LoadFunc
		expectedError error
		expectedVal   string
		expectedKVs   map[string]string
	}{
		{
			name:          "cache hit",
			cache:         NewLoadingCache(),
			key:           "2",
			load:          nil,
			expectedError: nil,
			expectedVal:   "two",
			expectedKVs:   map[string]string{"1": "one", "2": "two"},
		},
		{
			name:  "cache miss, loading error",
			cache: NewLoadingCache(),
			key:   "3",
			load: func(key interface{}) (interface{}, error) {
				return nil, errors.New("test error")
			},
			expectedError: errors.New("test error"),
			expectedVal:   "",
			expectedKVs:   map[string]string{"1": "one", "2": "two"},
		},
		{
			name:  "cache miss, loading okay, no eviction",
			cache: NewLoadingCache(),
			key:   "3",
			load: func(key interface{}) (interface{}, error) {
				return "three", nil
			},
			expectedError: nil,
			expectedVal:   "three",
			expectedKVs:   map[string]string{"1": "one", "2": "two", "3": "three"},
		},
		{
			name:  "cache miss, loading okay, eviction",
			cache: NewLoadingCache(2),
			key:   "3",
			load: func(key interface{}) (interface{}, error) {
				return "three", nil
			},
			expectedError: nil,
			expectedVal:   "three",
			expectedKVs:   map[string]string{"2": "two", "3": "three"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Populate the cache with initial key/value pairs.
			initKVs := []string{"1", "one", "2", "two"}
			for i := 0; i < len(initKVs)/2; i++ {
				test.cache.cache.Add(initKVs[i*2], initKVs[i*2+1])
			}
			val, err := test.cache.Get(test.key, test.load)
			if test.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedVal, val.(string))
			}
			// Do post Get() call cache check.
			kvs := make(map[string]string)
			for _, k := range test.cache.cache.Keys() {
				v, found := test.cache.cache.Get(k)
				assert.True(t, found)
				kvs[k.(string)] = v.(string)
			}
			assert.Equal(t, test.expectedKVs, kvs)
		})
	}
}

const (
	benchRandSeed = 123
	benchKeySize  = 10000
)

func getBenchKey() string {
	return "/AAA/BBB/CCC/DDD[@price = " + strconv.Itoa(rand.Intn(benchKeySize)) + " or . != 'd']"
}

func BenchmarkLoadingCache_NoCapacityLimit(b *testing.B) {
	rand.Seed(benchRandSeed)
	cache := NewLoadingCache()
	for i := 0; i < b.N; i++ {
		_, err := cache.Get(getBenchKey(), func(key interface{}) (interface{}, error) {
			return xpath.Compile(key.(string))
		})
		assert.NoError(b, err)
	}
}

func BenchmarkLoadingCache_SmallCapacityLimit(b *testing.B) {
	rand.Seed(benchRandSeed)
	cache := NewLoadingCache(benchKeySize / 100)
	for i := 0; i < b.N; i++ {
		_, err := cache.Get(getBenchKey(), func(key interface{}) (interface{}, error) {
			return xpath.Compile(key.(string))
		})
		assert.NoError(b, err)
	}
}
func BenchmarkLoadingCache_LargeCapacityLimit(b *testing.B) {
	rand.Seed(benchRandSeed)
	cache := NewLoadingCache(benchKeySize / 2)
	for i := 0; i < b.N; i++ {
		_, err := cache.Get(getBenchKey(), func(key interface{}) (interface{}, error) {
			return xpath.Compile(key.(string))
		})
		assert.NoError(b, err)
	}
}
