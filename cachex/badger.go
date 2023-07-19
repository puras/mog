package cachex

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"strings"
	"time"
	"unsafe"
)

type BadgerConfig struct {
	Path string
}

func NewBadgerCache(cfg BadgerConfig, opts ...Option) Cache {
	defaultOpts := &options{
		Delimiter: defaultDelimiter,
	}

	for _, o := range opts {
		o(defaultOpts)
	}

	badgerOpts := badger.DefaultOptions(cfg.Path)
	badgerOpts = badgerOpts.WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(badgerOpts)
	if err != nil {
		panic(err)
	}

	return &badgerCache{
		opts: defaultOpts,
		db:   db,
	}
}

type badgerCache struct {
	opts *options
	db   *badger.DB
}

func (o *badgerCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	return o.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(o.strToBytes(o.getKey(ns, key)), o.strToBytes(value))
		if len(expiration) > 0 {
			entry = entry.WithTTL(expiration[0])
		}
		return txn.SetEntry(entry)
	})
}

func (o *badgerCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	value := ""
	ok := false
	err := o.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(o.strToBytes(o.getKey(ns, key)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}
		ok = true
		val, err := item.ValueCopy(nil)
		value = o.bytesToStr(val)
		return err
	})
	if err != nil {
		return "", false, err
	}
	return value, ok, nil
}

func (o *badgerCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	value, ok, err := o.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}

	err = o.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(o.strToBytes(o.getKey(ns, key)))
	})
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (o *badgerCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	exists := false
	err := o.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(o.strToBytes(o.getKey(ns, key)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}
		exists = true
		return nil
	})
	return exists, err
}

func (o *badgerCache) Delete(ctx context.Context, ns, key string) error {
	b, err := o.Exists(ctx, ns, key)
	if err != nil {
		return err
	} else if !b {
		return nil
	}
	return o.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(o.strToBytes(o.getKey(ns, key)))
	})
}

func (o *badgerCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	return o.db.View(func(txn *badger.Txn) error {
		iterOpts := badger.DefaultIteratorOptions
		iterOpts.Prefix = o.strToBytes(o.getKey(ns, ""))
		it := txn.NewIterator(iterOpts)
		defer it.Close()

		it.Rewind()
		for it.Valid() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			key := o.bytesToStr(item.Key())
			if !fn(ctx, strings.TrimPrefix(key, o.getKey(ns, "")), o.bytesToStr(val)) {
				break
			}
			it.Next()
		}
		return nil
	})
}

func (o *badgerCache) Close(ctx context.Context) error {
	return o.db.Close()
}

func (o *badgerCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, o.opts.Delimiter, key)
}

func (o *badgerCache) strToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func (o *badgerCache) bytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
