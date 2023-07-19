package jwtx

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/puras/mog/config"
)

type Auther interface {
	GenerateToken(ctx context.Context, subject string) (TokenInfo, error)
	DestroyToken(ctx context.Context, accessToken string) error
	ParseSubject(ctx context.Context, accessToken string) (string, error)
	Release(ctx context.Context) error
}

func InitAuth(ctx context.Context) (Auther, func(), error) {
	cfg := config.C.Middleware.Auth
	var opts []Option
	opts = append(opts)
	fmt.Println(cfg)
	return nil, nil, nil
}

const defaultKey = ""

type options struct {
	signingMethod jwt.SigningMethod
	signingKey    []byte
	keyFuncs      []func(*jwt.Token) (any, error)
	expired       int
	tokeyType     string
}

type Option func(*options)

func SetSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}
