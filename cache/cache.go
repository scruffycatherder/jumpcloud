package cache

import (
	"sync"
)

type HashCache struct {
	sync.RWMutex
	cache map[uint64]string
}

func NewHashCache() *HashCache {
	return &HashCache{
		cache: make(map[uint64]string),
	}
}

func (hashCache *HashCache) Put(key uint64, hash string) {
	hashCache.Lock()
	hashCache.cache[key] = hash
	hashCache.Unlock()
}

func (hashCache *HashCache) Get(key uint64) (hash string, ok bool) {
	hashCache.RLock()
	hash, ok = hashCache.cache[key]
	hashCache.RUnlock()
	return
}

func (hashCache *HashCache) Delete(key uint64) {
	hashCache.Lock()
	delete(hashCache.cache, key)
	hashCache.Unlock()
}
