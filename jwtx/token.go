package jwtx

import jsoniter "github.com/json-iterator/go"

type TokenInfo interface {
	GetAccessToken() string
	GetRefreshToken() string
	GetTokenType() string
	GetExpiresAt() int64
	EncodeToJSON() ([]byte, error)
}

type tokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int64  `json:"expires_at"`
}

func (o *tokenInfo) GetAccessToken() string {
	return o.AccessToken
}

func (o *tokenInfo) GetRefreshToken() string {
	return o.RefreshToken
}

func (o *tokenInfo) GetTokenType() string {
	return o.TokenType
}

func (o *tokenInfo) GetExpiresAt() int64 {
	return o.ExpiresAt
}

func (o *tokenInfo) EncodeToJSON() ([]byte, error) {
	return jsoniter.Marshal(o)
}
