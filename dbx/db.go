package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/puras/mog/config"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	sdmysql "github.com/go-sql-driver/mysql"
)

func InitDB(ctx context.Context) (*gorm.DB, func(), error) {
	cfg := config.C.Storage.DataBase
	if !cfg.Enable {
		return nil, nil, nil
	}
	resolver := make([]ResolverConfig, len(cfg.Resolver))
	for i, v := range cfg.Resolver {
		resolver[i] = ResolverConfig{
			DBType:   v.DBType,
			Sources:  v.Sources,
			Replicas: v.Replicas,
			Tables:   v.Tables,
		}
	}

	db, err := NewDB(Config{
		Debug:        cfg.Debug,
		DBType:       cfg.Type,
		DSN:          cfg.DSN,
		MaxLifetime:  cfg.MaxLifetime,
		MaxIdleTime:  cfg.MaxIdleTime,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
		TablePrefix:  cfg.TablePrefix,
		Resolver:     resolver,
	})

	if err != nil {
		return nil, nil, err
	}

	return db, func() {
		sqlDB, err := db.DB()
		if err != nil {
			_ = sqlDB.Close()
		}
	}, nil
}

type Trans struct {
	DB *gorm.DB
}

type TransFunc func(context.Context) error

type ResolverConfig struct {
	DBType   string
	Sources  []string
	Replicas []string
	Tables   []string
}

type Config struct {
	Debug        bool
	DBType       string
	DSN          string
	MaxLifetime  int
	MaxIdleTime  int
	MaxOpenConns int
	MaxIdleConns int
	TablePrefix  string
	Resolver     []ResolverConfig
}

func NewDB(c Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch strings.ToLower(c.DBType) {
	case "mysql":
		if err := createDataBaseWithMySQL(c.DSN); err != nil {
			return nil, err
		}
		dialector = mysql.Open(c.DSN)
	case "postgres":
		dialector = postgres.Open(c.DSN)
	case "sqlite3":
		_ = os.MkdirAll(filepath.Dir(c.DSN), os.ModePerm)
		dialector = sqlite.Open(c.DSN)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", c.DBType)
	}

	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   c.TablePrefix,
			SingularTable: true,
		},
		Logger: logger.Discard,
	}
	if c.Debug {
		config.Logger = logger.Default
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}
	if len(c.Resolver) > 0 {
		resolver := &dbresolver.DBResolver{}
		for _, r := range c.Resolver {
			cfg := dbresolver.Config{}

			var open func(dsn string) gorm.Dialector
			dbType := strings.ToLower(r.DBType)
			switch dbType {
			case "mysql":
				open = mysql.Open
			case "postgres":
				open = postgres.Open
			case "sqlite3":
				open = sqlite.Open
			default:
				continue
			}

			for _, replica := range r.Replicas {
				if dbType == "sqlite3" {
					_ = os.MkdirAll(filepath.Dir(c.DSN), os.ModePerm)
				}
				cfg.Replicas = append(cfg.Replicas, open(replica))
			}
			for _, source := range r.Sources {
				if dbType == "sqlite3" {
					_ = os.MkdirAll(filepath.Dir(c.DSN), os.ModePerm)
				}
				cfg.Sources = append(cfg.Sources, open(source))
			}
			tables := stringSliceToInterfaceSlice(r.Tables)
			resolver.Register(cfg, tables...)
			zap.L().Info(fmt.Sprintf("Use resolver, #tables: %v, #replicas: %v, #sources: %v \n",
				tables, r.Replicas, r.Sources))
		}
		resolver.SetMaxIdleConns(c.MaxIdleConns).
			SetMaxOpenConns(c.MaxOpenConns).
			SetConnMaxLifetime(time.Duration(c.MaxLifetime) * time.Second).
			SetConnMaxIdleTime(time.Duration(c.MaxIdleTime) * time.Second)
		if err := db.Use(resolver); err != nil {
			return nil, err
		}
	}
	if c.Debug {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.MaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(c.MaxIdleTime) * time.Second)

	return db, nil
}

func stringSliceToInterfaceSlice(s []string) []any {
	r := make([]any, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r
}

func createDataBaseWithMySQL(dsn string) error {
	cfg, err := sdmysql.ParseDSN(dsn)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/", cfg.User, cfg.Passwd, cfg.Addr))
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET = `utf8mb4`;", cfg.DBName)
	_, err = db.Exec(query)
	return err
}

func LikeParameter(v string) string {
	return "%" + v + "%"
}
