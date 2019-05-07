# Cache

A simple cache framework for golang.


## Usage


Init a cache:

```golang
import "github.com/sleagon/cache"

// simple memory cache middleware
type SimpleMemoryCache struct {
	data map[string]int64
}

// Set value to map in memory
func (c *SimpleMemoryCache) Set(ctx *Context, next Next) {
	c.data[ctx.Key] = ctx.Value.(int64)

	// call next to make sure all memory cache ready
	// maybe you should check next is nil or not
	next(ctx)
}

// Get get value from memory cache
func (c *SimpleMemoryCache) Get(ctx *Context, next Next) {
	// check key in c.data
	v, ex := c.data[ctx.Key]
	if ex {
		// reset value/err/source
		ctx.Value = v
		ctx.Err = nil
		ctx.Source = c.Source()
		return
	}
	// find in chained caches
	next(ctx)
	// write back to memory cache
	if ctx.Err == nil && ctx.Value != nil {
		c.data[ctx.Key] = ctx.Value.(int64)
	}
}

// delete (expire) particular key
func (c *SimpleMemoryCache) Del(ctx *Context, next Next) {
	delete(c.data, ctx.Key)
}

// Source return the source name of this memory
func (c *SimpleMemoryCache) Source() string {
	return "simple-cache"
}

func main() {
  c := cache.New()
  simpleM := &SimpleMemoryCache{
    data: make(map[string]int64),
  }
  cache.Use(simpleM)
  now := time.Now().Unix()
  cache.Set("foo", now)
  cached, err := cache.Get("foo")

  println(cached, err)
}

```