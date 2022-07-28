package mog

import (
	"fmt"

	"gorm.io/gorm/logger"

	"gorm.io/gorm/schema"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/**
* @project kudo
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-18 22:12
 */
var db *gorm.DB

func DB() *gorm.DB {
	InitDB()
	return db
}

func InitDB() {
	if db == nil {
		initDB()
	}
}

func CloseDB() {
	if db != nil {
		logrus.Infof("其实想关闭，但是新版好像修改了，回头再调")
	}
}

func initDB() {
	config := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		viper.GetString("db.username"),
		viper.GetString("db.passwd"),
		viper.GetString("db.location"),
		viper.GetString("db.port"),
		viper.GetString("db.name"),
		true,
		"Local",
	)
	fmt.Println(fmt.Sprintf("[DB Config] %s", config))

	var infoLevel = logger.Warn
	if viper.GetString("runmode") == "debug" {
		infoLevel = logger.Info
	}

	var err error
	db, err = gorm.Open(mysql.Open(config), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   viper.GetString("db.table.prefix"),
			SingularTable: viper.GetBool("db.table.singular"),
		},
		Logger: logger.Default.LogMode(infoLevel),
	})
	if err != nil {
		logrus.Fatalf("Connecting data store failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatalf("DB() failed: %v", err)
	}
	// 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，
	// 可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(2000)
	// 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(1000)
	sqlDB.SetMaxIdleConns(0)
}
