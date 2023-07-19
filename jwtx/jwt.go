package jwtx

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/puras/mog/cachex"
	"github.com/puras/mog/config"
	"time"
)

type Auth interface {
	GenerateToken(ctx context.Context, subject string) (TokenInfo, error)
	DestroyToken(ctx context.Context, token string) error
	ParseSubject(ctx context.Context, token string) (string, error)
	Release(ctx context.Context) error
}

const defaultKey = "CG24SDVP8OHPK395GB5G"

var ErrInvalidToken = errors.New("Invalid token")

func InitAuth(ctx context.Context) (Auth, func(), error) {
	cfg := config.C.Middleware.Auth
	var opts []Option
	opts = append(opts, SetExpired(cfg.Expired))
	opts = append(opts, SetSigningKey(cfg.SigningKey))

	var method jwt.SigningMethod
	switch cfg.SigningMethod {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "HS384":
		method = jwt.SigningMethodHS384
	default:
		method = jwt.SigningMethodHS512
	}
	opts = append(opts, SetSigningMethod(method))

	var cache cachex.Cache
	switch cfg.Store.Type {
	case "redis":
		cache = cachex.NewRedisCache(cachex.RedisConfig{
			Addr:     cfg.Store.Redis.Addr,
			DB:       cfg.Store.Redis.DB,
			Username: cfg.Store.Redis.Username,
			Password: cfg.Store.Redis.Password,
		}, cachex.WithDelimiter(cfg.Store.Delimiter))
	case "badger":
		cache = cachex.NewBadgerCache(cachex.BadgerConfig{
			Path: cfg.Store.Badger.Path,
		}, cachex.WithDelimiter(cfg.Store.Delimiter))
	}

	auth := New(NewStoreWithCache(cache), opts...)
	return auth, func() {
		_ = auth.Release(ctx)
	}, nil
}

type options struct {
	signingMethod jwt.SigningMethod
	signingKey    []byte
	signingKey2   []byte
	keyFuncs      []func(*jwt.Token) (any, error)
	expired       int
	tokenType     string
}

type Option func(*options)

func SetSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

func SetSigningKey(key string) Option {
	return func(o *options) {
		o.signingKey = []byte(key)
	}
}

//func SetSigningKey(key, oldKey string) Option {
//	return func(o *options) {
//		o.signingKey = []byte(key)
//		if oldKey != "" && key != oldKey {
//			o.signingKey2 = []byte(oldKey)
//		}
//	}
//}

func SetExpired(expired int) Option {
	return func(o *options) {
		o.expired = expired
	}
}

func New(store Store, opts ...Option) Auth {
	o := options{
		tokenType:     "Bearer",
		expired:       7200,
		signingMethod: jwt.SigningMethodHS512,
		signingKey:    []byte(defaultKey),
	}

	for _, opt := range opts {
		opt(&o)
	}

	o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return o.signingKey, nil
	})

	return &jwtAuth{
		opts:  &o,
		store: store,
	}
}

type jwtAuth struct {
	opts  *options
	store Store
}

func (o *jwtAuth) GenerateToken(ctx context.Context, subject string) (TokenInfo, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(o.opts.expired) * time.Second).Unix()

	token := jwt.NewWithClaims(o.opts.signingMethod, &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt,
		NotBefore: now.Unix(),
		Subject:   subject,
	})

	tokenStr, err := token.SignedString(o.opts.signingKey)
	if err != nil {
		return nil, err
	}
	//err = o.callStore(func(store Store) error {
	//	return store.Set(ctx, tokenStr, time.Duration(expiresAt))
	//})
	//if err != nil {
	//	return nil, err
	//}
	tokenInfo := &tokenInfo{
		ExpiresAt:    expiresAt,
		TokenType:    o.opts.tokenType,
		AccessToken:  tokenStr,
		RefreshToken: tokenStr,
	}
	return tokenInfo, nil
}

func (o *jwtAuth) DestroyToken(ctx context.Context, token string) error {
	claims, err := o.parseToken(token)
	if err != nil {
		return err
	}
	return o.callStore(func(store Store) error {
		expired := time.Until(time.Unix(claims.ExpiresAt, 0))
		return store.Set(ctx, token, expired)
	})
}

func (o *jwtAuth) ParseSubject(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrInvalidToken
	}

	claims, err := o.parseToken(token)
	if err != nil {
		return "", err
	}

	err = o.callStore(func(store Store) error {
		if exists, err := store.Check(ctx, token); err != nil {
			return err
		} else if exists {
			return ErrInvalidToken
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

func (o *jwtAuth) Release(ctx context.Context) error {
	return o.callStore(func(store Store) error {
		return store.Close(ctx)
	})
}

func (o *jwtAuth) parseToken(token string) (*jwt.StandardClaims, error) {
	var (
		tk  *jwt.Token
		err error
	)

	for _, keyFunc := range o.opts.keyFuncs {
		tk, err = jwt.ParseWithClaims(token, &jwt.StandardClaims{}, keyFunc)
		if err != nil || tk == nil || !tk.Valid {
			continue
		}
		break
	}
	if err != nil || tk == nil || !tk.Valid {
		return nil, ErrInvalidToken
	}
	return tk.Claims.(*jwt.StandardClaims), nil
}

func (o *jwtAuth) callStore(fn func(Store) error) error {
	if store := o.store; store != nil {
		return fn(store)
	}
	return nil
}
