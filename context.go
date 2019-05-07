package cache

type Context struct {
	Key     string
	Value   interface{}
	Source  string
	Err     error
	payload map[string]interface{}
}

func (ctx *Context) Set(key string, value interface{}) {
	ctx.payload[key] = value
}

func (ctx *Context) Get(key string) interface{} {
	return ctx.payload[key]
}
