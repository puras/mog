package config

import (
	"fmt"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/pelletier/go-toml"
)

var (
	C    = new(Config)
	once sync.Once
)

func MustLoad(name string) {
	once.Do(func() {
		tree, err := toml.LoadFile(name)
		if err != nil {
			panic(fmt.Sprintf("Failed to load config file %s: %s", name, err.Error()))
		}
		if err = tree.Unmarshal(C); err != nil {
			panic(fmt.Sprintf("Failed to unmarshal config %s: %s", name, err.Error()))
		}
		if err = C.PreLoad(); err != nil {
			panic(fmt.Sprintf("Failed to preload config %s: %s", name, err.Error()))
		}
	})
}

type Config struct {
	General    General
	Storage    Storage
	Logger     Logger
	Middleware Middleware
	ExtConfig  any
}

type General struct {
	AppName           string `default:"mog"`
	DebugMode         bool
	ContextPath       string `default:""`
	PprofAddr         string
	EnableSwagger     bool
	EnablePrintConfig bool
	HTTP              struct {
		Addr            string `default:":8000"`
		ShutdownTimeout int    `default:"10"`
		ReadTimeout     int    `default:"60"` // seconds
		WriteTimeout    int    `default:"60"` // seconds
		IdleTimeout     int    `default:"10"` // seconds
		CertFile        string
		KeyFile         string
	}
}

type Storage struct {
	Cache struct {
		Type      string `default:"memory"` // memory/badger/redis
		Delimiter string `default:":"`
		Memory    struct {
			CleanupInterval int `default:"60"`
		}
		Badger struct {
			Path string
		}
		Redis struct {
			Addr     string
			Username string
			Password string
			DB       int
		}
	}
	DataBase struct {
		Enable       bool `default:"true"`
		Debug        bool
		Type         string `default:"sqlite3"` // sqlite3/mysql/postgres
		DSN          string `default:"data/sqlite/106hz.db"`
		MaxLifetime  int    `default:"86400"`
		MaxIdleTime  int    `default:"3600"`
		MaxOpenConns int    `default:"100"`
		MaxIdleConns int    `default:"50"`
		TablePrefix  string `default:""`
		AutoMigrate  bool
		Resolver     []struct {
			DBType   string   // sqlite3/mysql/postgres
			Sources  []string // DSN
			Replicas []string // DSN
			Tables   []string
		}
	}
}

type Logger struct {
	Debug      bool
	Level      string // debug/info/warn/error/dpanic/panic/fatal
	CallerSkip int
	File       struct {
		Enable     bool
		Path       string
		MaxSize    int
		MaxBackups int
	}
}

type Middleware struct {
	Recovery struct {
		Skip int `default:"3"`
	}
	CORS struct {
		Enable                 bool
		AllowAllOrigins        bool
		AllowOrigins           []string
		AllowMethods           []string
		AllowHeaders           []string
		AllowCredentials       bool
		ExposeHeaders          []string
		MaxAge                 int
		AllowWildcard          bool
		AllowBrowserExtensions bool
		AllowWebSockets        bool
		AllowFiles             bool
	}
	Trace struct {
		SkippedPathPrefixes []string
		RequestHeaderKey    string `default:"X-Request-Id"`
		ResponseTraceKey    string `default:"X-Trace-Id"`
	}
	Logger struct {
		SkippedPathPrefixes      []string
		MaxOutputRequestBodyLen  int `default:"4096"`
		MaxOutputResponseBodyLen int `default:"1024"`
	}
	CopyBody struct {
		SkippedPathPrefixes []string
		MaxContentLen       int64 `default:33554432` // max content length (default 32MB)
	}
	RateLimiter struct {
		Enable              bool
		SkippedPathPrefixes []string
		Period              int // seconds
		MaxRequestsPerIP    int
		MaxRequestsPerUser  int
		Store               struct {
			Type   string // memory/redis
			Memory struct {
				Expiration      int `default:"3600"` // seconds
				CleanupInterval int `default:"60"`   // seconds
			}
			Redis struct {
				Addr     string
				Username string
				Password string
				DB       int
			}
		}
	}
	Auth struct {
		Disable             bool
		SkippedPathPrefixes []string
		SigningMethod       string `default:"HS512"` // HS256/HS384/HS512
		SigningKey          string `default:"cptbtptpbcptdtptp"`
		Expired             int    `default:"86400"`
		Store               struct {
			Type      string `default:"badger"` // badger/redis
			Delimiter string `default:":"`      // delimiter for key
			Badger    struct {
				Path string `default:"data/auth"`
			}
			Redis struct {
				Addr     string
				Username string
				Password string
				DB       int
			}
		}
	}
}

func (c *Config) IsDebug() bool {
	return c.General.DebugMode
}

func (c *Config) String() string {
	b, err := jsoniter.MarshalIndent(c, "", "  ")
	if err != nil {
		panic("Failed to marshal config: " + err.Error())
	}
	return string(b)
}

func (c *Config) PreLoad() error {
	if addr := c.Storage.Cache.Redis.Addr; addr != "" {
		username := c.Storage.Cache.Redis.Username
		password := c.Storage.Cache.Redis.Password
		if c.Middleware.RateLimiter.Store.Redis.Addr == "" {
			c.Middleware.RateLimiter.Store.Redis.Addr = addr
			c.Middleware.RateLimiter.Store.Redis.Username = username
			c.Middleware.RateLimiter.Store.Redis.Password = password
		}
		if c.Middleware.Auth.Store.Type == "redis" &&
			c.Middleware.Auth.Store.Redis.Addr == "" {
			c.Middleware.Auth.Store.Redis.Addr = addr
			c.Middleware.Auth.Store.Redis.Username = username
			c.Middleware.Auth.Store.Redis.Password = password
		}
	}
	return nil
}

func (c *Config) Print() {
	fmt.Println("// -------------------- Load configurations start --------------------")
	fmt.Println(c.String())
	fmt.Println("// -------------------- Load configurations end --------------------")
}
