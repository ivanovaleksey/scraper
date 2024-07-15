package scraper

import (
	"sync"
)

type Cache struct {
	memo map[string]struct{}
	mux  sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		memo: make(map[string]struct{}),
	}
}

func (c *Cache) Set(url string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.memo[url] = struct{}{}
}

func (c *Cache) SetMulti(urls ...string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	for _, url := range urls {
		c.memo[url] = struct{}{}
	}
}

func (c *Cache) Has(url string) bool {
	c.mux.RLock()
	defer c.mux.RUnlock()
	_, ok := c.memo[url]
	return ok
}

func (c *Cache) GetAll() []string {
	c.mux.RLock()
	defer c.mux.RUnlock()

	items := make([]string, 0, len(c.memo))
	for url := range c.memo {
		items = append(items, url)
	}
	return items
}
