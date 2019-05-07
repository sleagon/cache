# Cache

A simple cache framework for golang.
## Features

In order to keep our system as strong as possible, we may build multi level cache in production environment, such as memory/file system/redis/oss... This package is written to manage this kind of problem. Just add your cacher middleware to the cache chain like koa middleware, then use cache.get/set simply. 😁😁😁

 - Koa-like cache handler.
 - Multi level cacher supported.
 - Simple, Reliable, Flexible.

## Usage


A simple memory cache:

```golang
package main

import (
	"time"

	"github.com/sleagon/cache"
)

// simple memory cache middleware
type SimpleMemoryCache struct {
	data map[string]int64
}

// Set value to map in memory
func (c *SimpleMemoryCache) Set(ctx *cache.Context, next cache.Next) {
	c.data[ctx.Key] = ctx.Value.(int64)

	// call next to make sure all memory cache ready
	// maybe you should check next is nil or not
	next(ctx)
}

// Get get value from memory cache
func (c *SimpleMemoryCache) Get(ctx *cache.Context, next cache.Next) {
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
func (c *SimpleMemoryCache) Del(ctx *cache.Context, next cache.Next) {
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
	c.Use(simpleM)
	now := time.Now().Unix()
	c.Set("foo", now)
	cached, err := c.Get("foo")

	println(cached.(int64), err == nil)
}
```

## Cache Middleware

All you need is implement Set/Get/Del/Source, the framework will do all left:

```golang
// MockSource is a mocked database, like mysql or redis
type MockSource struct {
	data map[string]int64
}

// Set value to mysql
func (c *MockSource) Set(ctx *Context, next Next) {
	c.data[ctx.Key] = ctx.Value.(int64)
	// set it to memory cache, you may not call this if it's not needed.
	next(ctx)
}

// Get get value from mysql, no next called
func (c *MockSource) Get(ctx *Context, next Next) {
	v, ex := c.data[ctx.Key]
	if ex {
		ctx.Source = c.Source()
		ctx.Value = v
	} else {
		ctx.Err = errors.New("Data not exist in mock database")
	}
}

// Del delete some key from mysql, empty function.
func (c *MockSource) Del(ctx *Context, next Next) {
	// source cache, like mysql/oss/rdb, etc
	// do not delete the data in source cache
}

// Source like the name.
func (c *MockSource) Source() string {
	return "mock-source"
}
```