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

// TODO(bobl): There seems to be some inconsistency in the way err or ok
// are returned in tuples from function calls (e.g. map lookups vs. i/o reads
// and writes.  Investigate what the generally accepted standard is, and use it.

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
