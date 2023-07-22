package cachex

import (
	"context"
	"fmt"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

type Cache interface {
	Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error
	Get(ctx context.Context, ns, key string) (string, bool, error)
	GetAndDelete(ctx context.Context, ns, key string) (string, bool, error)
	Exists(ctx context.Context, ns, key string) (bool, error)
	Delete(ctx context.Context, ns, key string) error
	Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error
	Close(ctx context.Context) error
}

var defaultDelimiter = ":"

type options struct {
	Delimiter string
}

type Option func(*options)

func WithDelimiter(delimiter string) Option {
	return func(o *options) {
		o.Delimiter = delimiter
	}
}

type MemoryConfig struct {
	CleanupInterval time.Duration
}

func NewMemoryCache(cfg MemoryConfig, opts ...Option) Cache {
	defaultOpts := &options{
		Delimiter: defaultDelimiter,
	}

	for _, o := range opts {
		o(defaultOpts)
	}

	return &memCache{
		opts:  defaultOpts,
		cache: cache.New(0, cfg.CleanupInterval),
	}
}

type memCache struct {
	opts  *options
	cache *cache.Cache
}

func (o *memCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, o.opts.Delimiter, key)
}

func (o *memCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	o.cache.Set(o.getKey(ns, key), value, exp)
	return nil
}

func (o *memCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	val, ok := o.cache.Get(o.getKey(ns, key))
	if !ok {
		return "", false, nil
	}
	return val.(string), ok, nil
}

func (o *memCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	val, ok, err := o.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}
	o.cache.Delete(o.getKey(ns, key))
	return val, true, nil
}

func (o *memCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	_, ok := o.cache.Get(o.getKey(ns, key))
	return ok, nil
}

func (o *memCache) Delete(ctx context.Context, ns, key string) error {
	o.cache.Delete(o.getKey(ns, key))
	return nil
}

func (o *memCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	for k, v := range o.cache.Items() {
		if strings.HasPrefix(k, o.getKey(ns, "")) {
			if !fn(ctx, strings.TrimPrefix(k, o.getKey(ns, "")), v.Object.(string)) {
				break
			}
		}
	}
	return nil
}

func (o *memCache) Close(ctx context.Context) error {
	o.cache.Flush()
	return nil
}
