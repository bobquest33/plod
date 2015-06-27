package dao

import "sync"

var (
	DefaultCache Cache
)

func init() {
	DefaultCache = NewMemoryCache()
}

type MemoryCache struct {
	c  map[string]struct{} // Empty struct takes up no memory
	mu sync.RWMutex        // Lock around concurrenct access to cache
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		c:  make(map[string]struct{}),
		mu: sync.RWMutex{},
	}
}

func (mc *MemoryCache) HaveVisited(url string) bool {
	mc.mu.RLock()
	_, present := mc.c[url]
	mc.mu.RUnlock()
	return present
}

func (mc *MemoryCache) Set(url string) {
	mc.mu.Lock()
	mc.c[url] = struct{}{}
	mc.mu.Unlock()
}
