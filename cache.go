package cache

import "errors"

// composed function type
type Next func(ctx *Context)

// function next has no second param
type handler func(ctx *Context, n Next)

// compose handler function array into one single next function
func compose(hs []handler) Next {
	return func(ctx *Context) {
		if len(hs) > 0 {
			hs[0](ctx, compose(hs[1:]))
		}
	}
}

func reverse(hs []handler) []handler {
	l := len(hs)
	for k := 0; k < l/2; k++ {
		hs[k], hs[l-k-1] = hs[l-k-1], hs[k]
	}
	return hs
}

// Middleware all middlewares should have get/set/del
type Middleware interface {
	Get(ctx *Context, n Next)
	Set(ctx *Context, n Next)
	Del(ctx *Context, n Next)
	Source() string
}

// Cache is core struct of this package
type Cache struct {
	handlerMap map[string][]handler
	get        Next
	set        Next
	del        Next
}

// Use Add a new middleware to cache
func (c *Cache) Use(m Middleware) *Cache {
	get := append(c.handlerMap["get"], m.Get)
	set := append(c.handlerMap["set"], m.Set)
	del := append(c.handlerMap["del"], m.Del)
	c.handlerMap["get"] = get
	c.handlerMap["set"] = set
	c.handlerMap["del"] = del
	c.get = compose(get)
	c.set = compose(reverse(set))
	c.del = compose(reverse(del))
	return c
}

func New() *Cache {
	c := new(Cache)
	c.handlerMap = make(map[string][]handler)
	c.handlerMap["get"] = []handler{}
	c.handlerMap["set"] = []handler{}
	c.handlerMap["del"] = []handler{}
	return c
}

// GetContext get data with context
func (c *Cache) GetContext(ctx *Context) {
	if ctx.Key == "" {
		ctx.Err = errors.New("Key is required for Get")
		return
	}
	c.get(ctx)
}

// SetContext set data with context
func (c *Cache) SetContext(ctx *Context) {
	if ctx.Key == "" || ctx.Value == nil {
		ctx.Err = errors.New("Key and Value are required for Set")
		return
	}
	c.set(ctx)
}

// DelContext delete data with context
func (c *Cache) DelContext(ctx *Context) {
	if ctx.Key == "" {
		ctx.Err = errors.New("Key is required for Del")
		return
	}
	c.del(ctx)
}

// Get get value with key
func (c *Cache) Get(key string) (interface{}, error) {
	ctx := &Context{
		Key: key,
	}
	c.GetContext(ctx)
	return ctx.Value, ctx.Err
}

// Set set key:value
func (c *Cache) Set(key string, value interface{}) error {
	ctx := &Context{
		Key:   key,
		Value: value,
	}
	c.SetContext(ctx)
	return ctx.Err
}

// Del delete one key
func (c *Cache) Del(key string) error {
	ctx := &Context{
		Key: key,
	}
	c.SetContext(ctx)
	return ctx.Err
}
