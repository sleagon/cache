package cache

import (
	"errors"
	"testing"
	"time"
)

// Simple cache memory
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

// Test a simple cache
func TestCache(t *testing.T) {
	cache := New()
	simpleM := &SimpleMemoryCache{
		data: make(map[string]int64),
	}
	cache.Use(simpleM)
	now := time.Now().Unix()

	cache.Set("foo", now)
	cached, err := cache.Get("foo")

	if err != nil {
		t.Errorf("Failed to get data, err = %s", err)
	}

	if cached != now {
		t.Errorf("Failed to get, %d != %d", cached, now)
	}
}

// test a simple cache with GetContext
func TestCacheGetContext(t *testing.T) {
	cache := New()
	simpleM := &SimpleMemoryCache{
		data: make(map[string]int64),
	}
	cache.Use(simpleM)
	now := time.Now().Unix()

	cache.Set("foo", now)
	ctx := &Context{
		Key: "foo",
	}

	cache.GetContext(ctx)

	if ctx.Err != nil {
		t.Errorf("Failed to get data, err = %s", ctx.Err)
	}

	if ctx.Value != now {
		t.Errorf("Failed to get, %d != %d", ctx.Value, now)
	}

	if ctx.Source != "simple-cache" {
		t.Errorf("Get data from wrong source, %s", ctx.Source)
	}
}

// Test a simple cahined cache
func TestChainedCache(t *testing.T) {
	cache := New()
	simpleM := &SimpleMemoryCache{
		data: make(map[string]int64),
	}
	sourceM := &MockSource{
		data: make(map[string]int64),
	}
	var sample int64 = 20190102
	sourceM.data["foo"] = sample
	cache.Use(simpleM).Use(sourceM)

	// test normal data
	cached, err := cache.Get("foo")

	if err != nil {
		t.Errorf("Failed to get data, err = %s", err)
	}

	if cached != sample {
		t.Errorf("Failed to get, %d != %d", cached, sample)
	}

	// test non-existent key
	cached, err = cache.Get("bar")

	if err == nil {
		t.Error("Non-existent should return error from source database.")
	}

	// test invalid key
	cached, err = cache.Get("")
	if cached == nil {
		t.Error("Invalid key should failed.")
	}

	// set and get
	sample = 19700101
	cache.Set("bar", sample)
	cached, err = cache.Get("")
	if cached != sample {
		t.Errorf("Failed to set and get, %d != %d", sample, cached)
	}
}

func TestChainedSource(t *testing.T) {
	cache := New()
	simpleM := &SimpleMemoryCache{
		data: make(map[string]int64),
	}
	sourceM := &MockSource{
		data: make(map[string]int64),
	}
	var sample int64 = 20190102
	sourceM.data["foo"] = sample
	cache.Use(simpleM).Use(sourceM)

	ctx := &Context{
		Key: "foo",
	}
	cache.GetContext(ctx)

	if ctx.Err != nil {
		t.Error("[1] Failed to get data from mock database.")
	}

	if ctx.Value != sample {
		t.Errorf("[1] Got wrong value from mock database, %d != %d.", ctx.Value, sample)
	}

	if ctx.Source != "mock-source" {
		t.Errorf("[1] Got wrong source %s != mock-source", ctx.Source)
	}

	// set and get
	ctx = &Context{
		Key: "foo",
	}
	cache.GetContext(ctx)

	if ctx.Err != nil {
		t.Error("[2] Failed to get data from mock database.")
	}

	if ctx.Value != sample {
		t.Errorf("[2] Got wrong value from mock database, %d != %d.", ctx.Value, sample)
	}

	if ctx.Source != "simple-cache" {
		t.Errorf("[2] Got wrong source %s != simple-cache", ctx.Source)
	}

	cache.Use(simpleM).Use(sourceM)

	ctx = &Context{
		Key:   "bar",
		Value: int64(999),
	}
	cache.SetContext(ctx)

	if ctx.Err != nil || sourceM.data["bar"] != 999 {
		t.Errorf("Failed to set data to mock-database, %d.", sourceM.data["bar"])
	}

	if simpleM.data["bar"] != 999 {
		t.Errorf("Failed to set to simple cache, %d.", simpleM.data["bar"])
	}
}
