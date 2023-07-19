package cachex

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

func NewRedisCache(cfg RedisConfig, opts ...Option) Cache {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return newRedisCache(cli, opts...)
}

func NewRedisCacheWithClient(cli *redis.Client, opts ...Option) Cache {
	return newRedisCache(cli, opts...)
}

func NewRedisCacheWithClusterClient(cli *redis.ClusterClient, opts ...Option) Cache {
	return newRedisCache(cli, opts...)
}

func newRedisCache(cli redisClient, opts ...Option) Cache {
	defaultOpts := &options{
		Delimiter: defaultDelimiter,
	}

	for _, o := range opts {
		o(defaultOpts)
	}

	return &redisCache{
		opts: defaultOpts,
		cli:  cli,
	}
}

type redisClient interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Close() error
}

type redisCache struct {
	opts *options
	cli  redisClient
}

func (o *redisCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, o.opts.Delimiter, key)
}

func (o *redisCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	cmd := o.cli.Set(ctx, o.getKey(ns, key), value, exp)
	return cmd.Err()
}

func (o *redisCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	cmd := o.cli.Get(ctx, o.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return "", false, err
	}
	return cmd.Val(), true, nil
}

func (o *redisCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	value, ok, err := o.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}
	cmd := o.cli.Del(ctx, o.getKey(ns, key))
	if err := cmd.Err(); err != nil && err != redis.Nil {
		return "", false, err
	}
	return value, true, nil
}

func (o *redisCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	cmd := o.cli.Exists(ctx, o.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

func (o *redisCache) Delete(ctx context.Context, ns, key string) error {
	b, err := o.Exists(ctx, ns, key)
	if err != nil {
		return err
	} else if !b {
		return nil
	}
	cmd := o.cli.Del(ctx, o.getKey(ns, key))
	if err := cmd.Err(); err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (o *redisCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	var cursor uint64 = 0
LB_LOOP:
	for {
		cmd := o.cli.Scan(ctx, cursor, o.getKey(ns, "*"), 100)
		if err := cmd.Err(); err != nil {
			return err
		}
		keys, c, err := cmd.Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			cmd := o.cli.Get(ctx, key)
			if err := cmd.Err(); err != nil {
				if err == redis.Nil {
					continue
				}
				return err
			}
			if next := fn(ctx, strings.TrimPrefix(key, o.getKey(ns, "")), cmd.Val()); !next {
				break LB_LOOP
			}
		}
		if c == 0 {
			break
		}
		cursor = c
	}
	return nil
}

func (o *redisCache) Close(ctx context.Context) error {
	return o.cli.Close()
}
