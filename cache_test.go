package cache

import (
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	mp := make(map[string]string)
	mp["hello"] = "world"
	simple := func(pld *Payload, next Next) {
		body := mp[pld.Key()]
		if body != "" {
			pld.Body = body
			pld.Source = "simple"
		}
	}
	cc := NewCache()
	cc.Use(simple)

	if cc.Get("hello") != "world" {
		t.Error("Failed to get data from simple memory cache.")
	}

	if cc.Get("world") != nil {
		t.Error("Got wrong data from memory cache.")
	}
}

func TestChainedSimple(t *testing.T) {
	simpleGen := func(name string) Middleware {
		mp := make(map[string]string)
		mp[name] = strings.ToUpper(name)
		return func(pld *Payload, next Next) {
			body := mp[pld.Key()]
			if body != "" {
				pld.Body = mp[pld.Key()]
				pld.Source = name
				return
			}
			next(pld)
			if pld.Body != nil {
				mp[pld.Key()] = pld.Body.(string)
			}
		}
	}

	cc := NewCache()
	cc.Use(simpleGen("foo")).Use(simpleGen("bar"))

	if cc.Get("foo") != "FOO" {
		t.Error("Failed to get data from simple memory cache.")
	}
	if cc.Get("bar") != "BAR" {
		t.Error("Failed to get data from simple memory cache.")
	}
	if cc.Get("hello") != nil {
		t.Error("Got wrong value from chained cache.")
	}
}

func TestGetPayload(t *testing.T) {
	simpleGen := func(name string) Middleware {
		mp := make(map[string]string)
		mp[name] = strings.ToUpper(name)
		return func(pld *Payload, next Next) {
			body := mp[pld.Key()]
			if body != "" {
				pld.Body = mp[pld.Key()]
				pld.Source = name
				return
			}
			next(pld)
			if pld.Body != nil {
				mp[pld.Key()] = pld.Body.(string)
			}
		}
	}
	cc := NewCache()
	cc.Use(simpleGen("foo")).Use(simpleGen("bar"))

	pld := NewPld("bar", false)
	cc.GetPayload(pld)
	if pld.Body != "BAR" {
		t.Errorf("Failed to get payload %s != BAR", pld.Body)
	}

	if pld.Source != "bar" {
		t.Errorf("Got data from wrong source %s != bar", pld.Source)
	}

	// second time
	pld = NewPld("bar", false)
	cc.GetPayload(pld)
	if pld.Body != "BAR" {
		t.Errorf("Failed to get cached payload %s != BAR", pld.Body)
	}

	if pld.Source != "foo" {
		t.Errorf("Failed to cache FOO to simpl1 %s != foo", pld.Source)
	}
}

func TestForce(t *testing.T) {

	mp1 := make(map[string]string)
	mp1["test"] = "cached-data"
	mw1 := func(pld *Payload, next Next) {
		body := mp1[pld.Key()]
		if !pld.Force() && body != "" {
			pld.Body = body
			pld.Source = "mw1"
			return
		}
		next(pld)
		if pld.Body != nil {
			mp1[pld.Key()] = pld.Body.(string)
		}
	}

	mp2 := make(map[string]string)
	mp2["test"] = "force-data"
	mw2 := func(pld *Payload, next Next) {
		body := mp2[pld.Key()]
		if body != "" {
			pld.Body = body
		}
	}

	cc := NewCache()
	cc.Use(mw1)
	cc.Use(mw2)

	pld := NewPld("test", false)

	cc.GetPayload(pld)

	if pld.Body != "cached-data" {
		t.Errorf("Failed to get cached data %s != cached-data", pld.Body)
	}

	pld = NewPld("test", true)
	cc.GetPayload(pld)

	if pld.Body != "force-data" {
		t.Errorf("Failed to get force data %s != force-data", pld.Body)
	}

}
