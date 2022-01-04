package middleware

/**
 * @project momo-backend
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-24 22:35
 * @desc
 */
import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type MapClaims map[string]interface{}

type JwtMiddleware struct {
	Realm                 string
	SigningAlgorithm      string
	Key                   []byte
	Timeout               time.Duration
	MaxRefresh            time.Duration
	Authenticator         func(c *gin.Context) (interface{}, error)
	Authorizator          func(data interface{}, c *gin.Context) bool
	PayloadFunc           func(data interface{}) MapClaims
	Unauthorized          func(c *gin.Context, code int, message string)
	LoginResponse         func(c *gin.Context, code int, message string, time time.Time)
	LogoutResponse        func(c *gin.Context, code int)
	RefreshResponse       func(c *gin.Context, code int, message string, time time.Time)
	IdentityHandler       func(c *gin.Context) interface{}
	IdentityKey           string
	TokenLookup           string
	TokenHeadName         string
	TimeFunc              func() time.Time
	HTTPStatusMessageFunc func(e error, c *gin.Context) string
	PrivKeyFile           string
	PrivKeyBytes          []byte
	PubKeyFile            string
	PubKeyBytes           []byte
	privKey               *rsa.PrivateKey
	pubKey                *rsa.PublicKey
	SendCookie            bool
	CookieMaxAge          time.Duration
	SecureCookie          bool
	CookieHTTPOnly        bool
	CookieDomain          string
	SendAuthorization     bool
	DisabledAbort         bool
	CookieName            string
	CookieSameSite        http.SameSite
}

var (
	ErrMissingSecretKey         = errors.New("secret key is required")
	ErrForbidden                = errors.New("you don't have permission to access this resource")
	ErrMissingAuthenticatorFunc = errors.New("JWTMiddleware.Authenticator func is undefined")
	ErrMissingLoginValues       = errors.New("missing Username or Password")
	ErrFailedAuthentication     = errors.New("incorrect Username or Password")
	ErrFailedTokenCreation      = errors.New("failed to create JWT token")
	ErrExpiredToken             = errors.New("token is expired")
	ErrEmptyAuthHeader          = errors.New("auth header is empty")
	ErrMissingExpField          = errors.New("missing exp field")
	ErrWrongFormatOfExp         = errors.New("exp must be float64 format")
	ErrInvalidAuthHeader        = errors.New("auth header is invalid")
	ErrEmptyQueryToken          = errors.New("query token is empty")
	ErrEmptyCookieToken         = errors.New("cookie token is empty")
	ErrEmptyParamToken          = errors.New("parameter token is empty")
	ErrInvalidSigningAlgorithm  = errors.New("invalid signing algorithm")
	ErrNoPrivKeyFile            = errors.New("private key file unreadable")
	ErrNoPubKeyFile             = errors.New("public key file unreadable")
	ErrInvalidPrivKey           = errors.New("private key invalid")
	ErrInvalidPubKey            = errors.New("public key invalid")
	IdentityKey                 = "identity"
)

func NewJwtMiddleware(m *JwtMiddleware) (*JwtMiddleware, error) {
	if err := m.MiddlewareInit(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *JwtMiddleware) MiddlewareInit() error {
	if m.TokenLookup == "" {
		m.TokenLookup = "header:Authorization"
	}
	if m.SigningAlgorithm == "" {
		m.SigningAlgorithm = "HS256"
	}
	if m.Timeout == 0 {
		m.Timeout = time.Hour
	}
	if m.TimeFunc == nil {
		m.TimeFunc = time.Now
	}
	m.TokenHeadName = strings.TrimSpace(m.TokenHeadName)
	if len(m.TokenHeadName) == 0 {
		m.TokenHeadName = "Bearer"
	}
	if m.Authorizator == nil {
		m.Authorizator = func(data interface{}, c *gin.Context) bool {
			return true
		}
	}
	if m.Unauthorized == nil {
		m.Unauthorized = func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		}
	}
	if m.LoginResponse == nil {
		m.LoginResponse = func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"code":   http.StatusOK,
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		}
	}
	if m.LogoutResponse == nil {
		m.LogoutResponse = func(c *gin.Context, code int) {
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
			})
		}
	}
	if m.RefreshResponse == nil {
		m.RefreshResponse = func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"code":   http.StatusOK,
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		}
	}
	if m.IdentityKey == "" {
		m.IdentityKey = IdentityKey
	}
	if m.IdentityHandler == nil {
		m.IdentityHandler = func(c *gin.Context) interface{} {
			claims := ExtractClaims(c)
			return claims[m.IdentityKey]
		}
	}
	if m.HTTPStatusMessageFunc == nil {
		m.HTTPStatusMessageFunc = func(e error, c *gin.Context) string {
			return e.Error()
		}
	}
	if m.Realm == "" {
		m.Realm = "gin jwt"
	}
	if m.CookieMaxAge == 0 {
		m.CookieMaxAge = m.Timeout
	}
	if m.CookieName == "" {
		m.CookieName = "jwt"
	}
	if m.usingPublicKeyAlgo() {
		return m.readKeys()
	}
	if m.Key == nil {
		return ErrMissingSecretKey
	}
	return nil
}

func (m *JwtMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.middlewareImpl(c)
	}
}

func (m *JwtMiddleware) middlewareImpl(c *gin.Context) {
	claims, err := m.GetClaimsFromJWT(c)
	if err != nil {
		m.unauthorized(c, http.StatusUnauthorized, m.HTTPStatusMessageFunc(err, c))
		return
	}
	if claims["exp"] == nil {
		m.unauthorized(c, http.StatusBadRequest, m.HTTPStatusMessageFunc(ErrMissingExpField, c))
		return
	}
	if _, ok := claims["exp"].(float64); !ok {
		m.unauthorized(c, http.StatusBadRequest, m.HTTPStatusMessageFunc(ErrWrongFormatOfExp, c))
		return
	}
	if int64(claims["exp"].(float64)) < m.TimeFunc().Unix() {
		m.unauthorized(c, http.StatusUnauthorized, m.HTTPStatusMessageFunc(ErrExpiredToken, c))
		return
	}
	c.Set("JWT_PAYLOAD", claims)
	identity := m.IdentityHandler(c)

	if identity != nil {
		c.Set(m.IdentityKey, identity)
	}

	if !m.Authorizator(identity, c) {
		m.unauthorized(c, http.StatusForbidden, m.HTTPStatusMessageFunc(ErrForbidden, c))
		return
	}
	c.Next()
}

func (m *JwtMiddleware) GetClaimsFromJWT(c *gin.Context) (MapClaims, error) {
	token, err := m.ParseToken(c)
	if err != nil {
		return nil, err
	}
	if m.SendAuthorization {
		if v, ok := c.Get("JWT_TOKEN"); ok {
			c.Header("Authorization", m.TokenHeadName+" "+v.(string))
		}
	}
	claims := MapClaims{}
	for key, value := range token.Claims.(jwt.MapClaims) {
		claims[key] = value
	}
	return claims, nil
}

func (m *JwtMiddleware) LoginHandler(c *gin.Context) {
	if m.Authenticator == nil {
		m.unauthorized(c, http.StatusInternalServerError, m.HTTPStatusMessageFunc(ErrMissingAuthenticatorFunc, c))
		return
	}
	data, err := m.Authenticator(c)
	if err != nil {
		m.unauthorized(c, http.StatusUnauthorized, m.HTTPStatusMessageFunc(err, c))
		return
	}
	token := jwt.New(jwt.GetSigningMethod(m.SigningAlgorithm))
	claims := token.Claims.(jwt.MapClaims)

	if m.PayloadFunc != nil {
		for key, value := range m.PayloadFunc(data) {
			claims[key] = value
		}
	}
	expire := m.TimeFunc().Add(m.Timeout)
	claims["exp"] = expire.Unix()
	claims["orig_iat"] = m.TimeFunc().Unix()
	tokenString, err := m.signedString(token)

	if err != nil {
		m.unauthorized(c, http.StatusUnauthorized, m.HTTPStatusMessageFunc(ErrFailedTokenCreation, c))
		return
	}

	if m.SendCookie {
		expireCookie := m.TimeFunc().Add(m.CookieMaxAge)
		maxage := int(expireCookie.Unix() - m.TimeFunc().Unix())

		if m.CookieSameSite != 0 {
			c.SetSameSite(m.CookieSameSite)
		}

		c.SetCookie(
			m.CookieName,
			tokenString,
			maxage,
			"/",
			m.CookieDomain,
			m.SecureCookie,
			m.CookieHTTPOnly,
		)
	}
	m.LoginResponse(c, http.StatusOK, tokenString, expire)
}

func (m *JwtMiddleware) LogoutHandler(c *gin.Context) {
	if m.SendCookie {
		if m.CookieSameSite != 0 {
			c.SetSameSite(m.CookieSameSite)
		}

		c.SetCookie(
			m.CookieName,
			"",
			-1,
			"/",
			m.CookieDomain,
			m.SecureCookie,
			m.CookieHTTPOnly,
		)
	}
	m.LogoutResponse(c, http.StatusOK)
}

func (m *JwtMiddleware) RefreshHandler(c *gin.Context) {
	tokenString, expire, err := m.RefreshToken(c)
	if err != nil {
		m.unauthorized(c, http.StatusUnauthorized, m.HTTPStatusMessageFunc(err, c))
		return
	}
	m.RefreshResponse(c, http.StatusOK, tokenString, expire)
}

func (m *JwtMiddleware) RefreshToken(c *gin.Context) (string, time.Time, error) {
	claims, err := m.CheckIfTokenExpire(c)
	if err != nil {
		return "", time.Now(), err
	}
	newToken := jwt.New(jwt.GetSigningMethod(m.SigningAlgorithm))
	newClaims := newToken.Claims.(jwt.MapClaims)

	for key := range claims {
		newClaims[key] = claims[key]
	}
	expire := m.TimeFunc().Add(m.Timeout)
	newClaims["exp"] = expire.Unix()
	newClaims["orig_iat"] = m.TimeFunc().Unix()
	tokenString, err := m.signedString(newToken)

	if err != nil {
		return "", time.Now(), err
	}

	if m.SendCookie {
		expireCookie := m.TimeFunc().Add(m.CookieMaxAge)
		maxage := int(expireCookie.Unix() - time.Now().Unix())

		if m.CookieSameSite != 0 {
			c.SetSameSite(m.CookieSameSite)
		}

		c.SetCookie(
			m.CookieName,
			tokenString,
			maxage,
			"/",
			m.CookieDomain,
			m.SecureCookie,
			m.CookieHTTPOnly,
		)
	}
	return tokenString, expire, nil
}

func (m *JwtMiddleware) CheckIfTokenExpire(c *gin.Context) (jwt.MapClaims, error) {
	token, err := m.ParseToken(c)

	if err != nil {
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok || validationErr.Errors != jwt.ValidationErrorExpired {
			return nil, err
		}
	}
	claims := token.Claims.(jwt.MapClaims)
	origIat := int64(claims["orig_iat"].(float64))

	if origIat < m.TimeFunc().Add(-m.MaxRefresh).Unix() {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

func (m *JwtMiddleware) TokenGenerator(data interface{}) (string, time.Time, error) {
	token := jwt.New(jwt.GetSigningMethod(m.SigningAlgorithm))
	claims := token.Claims.(jwt.MapClaims)

	if m.PayloadFunc != nil {
		for key, value := range m.PayloadFunc(data) {
			claims[key] = value
		}
	}

	expire := m.TimeFunc().UTC().Add(m.Timeout)
	claims["exp"] = expire.Unix()
	claims["orig_iat"] = m.TimeFunc().Unix()
	tokenString, err := m.signedString(token)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expire, nil
}

func (m *JwtMiddleware) ParseToken(c *gin.Context) (*jwt.Token, error) {
	var token string
	var err error

	methods := strings.Split(m.TokenLookup, ",")
	for _, method := range methods {
		if len(token) > 0 {
			break
		}
		parts := strings.Split(strings.TrimSpace(method), ":")
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case "header":
			token, err = m.jwtFromHeader(c, v)
		case "query":
			token, err = m.jwtFromQuery(c, v)
		case "cookie":
			token, err = m.jwtFromCookie(c, v)
		case "param":
			token, err = m.jwtFromParam(c, v)
		}
	}
	if err != nil {
		return nil, err
	}
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(m.SigningAlgorithm) != t.Method {
			return nil, ErrInvalidSigningAlgorithm
		}
		if m.usingPublicKeyAlgo() {
			return m.pubKey, nil
		}
		c.Set("JWT_TOKEN", token)
		return m.Key, nil
	})
}

func (m *JwtMiddleware) ParseTokenString(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(m.SigningAlgorithm) != t.Method {
			return nil, ErrInvalidSigningAlgorithm
		}
		if m.usingPublicKeyAlgo() {
			return m.pubKey, nil
		}
		return m.Key, nil
	})
}

func ExtractClaims(c *gin.Context) MapClaims {
	claims, exist := c.Get("JWT_PAYLOAD")
	if !exist {
		return make(MapClaims)
	}
	return claims.(MapClaims)
}

func ExtractClaimsFromToken(token *jwt.Token) MapClaims {
	if token == nil {
		return make(MapClaims)
	}
	claims := MapClaims{}
	for key, value := range token.Claims.(jwt.MapClaims) {
		claims[key] = value
	}
	return claims
}

func GetToken(c *gin.Context) string {
	token, exist := c.Get("JWT_TOKEN")
	if !exist {
		return ""
	}
	return token.(string)
}

func (m *JwtMiddleware) readKeys() error {
	err := m.privateKey()
	if err != nil {
		return err
	}
	err = m.publicKey()
	if err != nil {
		return err
	}
	return nil
}

func (m *JwtMiddleware) privateKey() error {
	var keyData []byte
	if m.PrivKeyFile == "" {
		keyData = m.PrivKeyBytes
	} else {
		filecontent, err := ioutil.ReadFile(m.PrivKeyFile)
		if err != nil {
			return ErrNoPrivKeyFile
		}
		keyData = filecontent
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return ErrInvalidPrivKey
	}
	m.privKey = key
	return nil
}

func (m *JwtMiddleware) publicKey() error {
	var keyData []byte
	if m.PubKeyFile == "" {
		keyData = m.PubKeyBytes
	} else {
		filecontent, err := ioutil.ReadFile(m.PubKeyFile)
		if err != nil {
			return ErrNoPubKeyFile
		}
		keyData = filecontent
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return ErrInvalidPubKey
	}
	m.pubKey = key
	return nil
}

func (m *JwtMiddleware) usingPublicKeyAlgo() bool {
	switch m.SigningAlgorithm {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (m *JwtMiddleware) unauthorized(c *gin.Context, code int, message string) {
	c.Header("WWW-Authenticate", "JWT realm="+m.Realm)
	if !m.DisabledAbort {
		c.Abort()
	}
	m.Unauthorized(c, code, message)
}

func (m *JwtMiddleware) jwtFromHeader(c *gin.Context, key string) (string, error) {
	authHeader := c.Request.Header.Get(key)

	if authHeader == "" {
		return "", ErrEmptyAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == m.TokenHeadName) {
		return "", ErrInvalidAuthHeader
	}
	return parts[1], nil
}

func (m *JwtMiddleware) jwtFromQuery(c *gin.Context, key string) (string, error) {
	token := c.Query(key)
	if token == "" {
		return "", ErrEmptyQueryToken
	}
	return token, nil
}

func (m *JwtMiddleware) jwtFromCookie(c *gin.Context, key string) (string, error) {
	cookie, _ := c.Cookie(key)
	if cookie == "" {
		return "", ErrEmptyCookieToken
	}
	return cookie, nil
}

func (m *JwtMiddleware) jwtFromParam(c *gin.Context, key string) (string, error) {
	token := c.Param(key)
	if token == "" {
		return "", ErrEmptyParamToken
	}
	return token, nil
}

func (m *JwtMiddleware) signedString(token *jwt.Token) (string, error) {
	var tokenString string
	var err error
	if m.usingPublicKeyAlgo() {
		tokenString, err = token.SignedString(m.privKey)
	} else {
		tokenString, err = token.SignedString(m.Key)
	}
	return tokenString, err
}
