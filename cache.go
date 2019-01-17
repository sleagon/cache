package cache

// cache a simple package for caching

// Payload Payload used throughout all middlewares
type Payload struct {
	Source string
	key    string
	force  bool
	Body   interface{}
}

// NewPld new a simple Payload pointer
func NewPld(key string, force bool) *Payload {
	return &Payload{"", key, force, nil}
}

// NewCache return a cache instance
func NewCache() Cache {
	return Cache{[]Middleware{}, func(c *Payload) {}}
}

// Force whether get data without using cache
func (c *Payload) Force() bool {
	return c.force
}

// Key key of Payload
func (c *Payload) Key() string {
	return c.key
}

// Next point to the next Middleware
type Next func(*Payload)

// Middleware middleware for cache
type Middleware func(*Payload, Next)

// Cache core struct of cache
type Cache struct {
	middlewares []Middleware
	composed    Next
}

// Use add a new middleware
func (c *Cache) Use(m Middleware) *Cache {
	c.middlewares = append(c.middlewares, m)
	c.composed = compose(c.middlewares)
	return c
}

// GetPayload get data with payload
func (c *Cache) GetPayload(pld *Payload) {
	c.composed(pld)
}

// Get get data of key
func (c *Cache) Get(key string) interface{} {
	pld := NewPld(key, false)
	c.GetPayload(pld)
	return pld.Body
}

// compose compose all middlewares to a single middleware
func compose(hs []Middleware) Next {
	return func(pld *Payload) {
		if len(hs) > 0 {
			hs[0](pld, compose(hs[1:]))
		}
	}
}
