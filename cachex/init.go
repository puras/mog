package cachex

import (
	"context"
	"github.com/puras/mog/config"
	"time"
)

func InitCache(ctx context.Context) (Cache, func(), error) {
	cfg := config.C.Storage.Cache

	var cache Cache

	switch cfg.Type {
	case "redis":
		cache = NewRedisCache(RedisConfig{
			Addr:     cfg.Redis.Addr,
			DB:       cfg.Redis.DB,
			Username: cfg.Redis.Username,
			Password: cfg.Redis.Password,
		}, WithDelimiter(cfg.Delimiter))
	case "badger":
		cache = NewBadgerCache(BadgerConfig{
			Path: (cfg.Badger.Path),
		}, WithDelimiter(cfg.Delimiter))
	default:
		cache = NewMemoryCache(MemoryConfig{
			CleanupInterval: time.Second * time.Duration(cfg.Memory.CleanupInterval),
		})
	}
	return cache, func() {
		_ = cache.Close(ctx)
	}, nil
}
