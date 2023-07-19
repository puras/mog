package jwtx

import (
	"context"
	"github.com/puras/mog/cachex"
	"time"
)

type Store interface {
	Set(ctx context.Context, token string, expiration time.Duration) error
	Delete(ctx context.Context, token string) error
	Check(ctx context.Context, token string) (bool, error)
	Close(ctx context.Context) error
}

type storeOptions struct {
	CacheNS string // default "jwt
}

type StoreOption func(*storeOptions)

func WithCacheNS(ns string) StoreOption {
	return func(o *storeOptions) {
		o.CacheNS = ns
	}
}

func NewStoreWithCache(cache cachex.Cache, opts ...StoreOption) Store {
	s := &storeImpl{
		c: cache,
		opts: &storeOptions{
			CacheNS: "jwt",
		},
	}
	for _, opt := range opts {
		opt(s.opts)
	}
	return s
}

type storeImpl struct {
	opts *storeOptions
	c    cachex.Cache
}

func (o *storeImpl) Set(ctx context.Context, token string, expiration time.Duration) error {
	return o.c.Set(ctx, o.opts.CacheNS, token, "", expiration)
}

func (o *storeImpl) Delete(ctx context.Context, token string) error {
	return o.c.Delete(ctx, o.opts.CacheNS, token)
}

func (o *storeImpl) Check(ctx context.Context, token string) (bool, error) {
	return o.c.Exists(ctx, o.opts.CacheNS, token)
}

func (o *storeImpl) Close(ctx context.Context) error {
	return o.c.Close(ctx)
}
