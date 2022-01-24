package mog

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/onrik/logrus/filename"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/**
* @project kuko
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-18 20:38
 */
type Config struct {
	Name string
}

func InitConfig(cfg string) error {
	c := Config{
		Name: cfg,
	}
	err := c.initConfig()
	if err != nil {
		return err
	}

	// 初始化日志包
	c.initLog()
	// 监控配置文件变化并热加载程序
	c.watchConfig()

	return nil
}

func (c Config) initLog() {
	logFile := viper.GetString("log.logger_file")
	logDir := filepath.Dir(logFile)
	err := MakeDir(logDir)
	if err != nil {
		logrus.Panic("make logger file errcode.", errors.WithStack(err))
	}
	bizLogFile := viper.GetString("biz_log.logger_file")
	bizLogDir := filepath.Dir(bizLogFile)
	err = MakeDir(bizLogDir)
	if err != nil {
		logrus.Panic("make business logger file errcode.", errors.WithStack(err))
	}
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	blWriter, err := rotatelogs.New(
		logFile+".%Y%m%d",
		rotatelogs.WithMaxAge(time.Hour*24*viper.GetDuration("biz_log.log_max_date")),          // 文件最大保存时间
		rotatelogs.WithRotationTime(time.Hour*24*viper.GetDuration("biz_log.log_rotate_date")), // 日志切割时间间隔
	)
	if err != nil {
		logrus.Panic("Config local file business logger errcode.", errors.WithStack(err))
	}
	bllfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: blWriter,
	}, &logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000", FieldMap: logrus.FieldMap{logrus.FieldKeyTime: "createTime", logrus.FieldKeyLevel: "logLevel"}})
	logrus.AddHook(bllfHook)

	writer, err := rotatelogs.New(
		logFile+".%Y%m%d",
		rotatelogs.WithMaxAge(time.Hour*24*viper.GetDuration("log.log_max_date")),          // 文件最大保存时间
		rotatelogs.WithRotationTime(time.Hour*24*viper.GetDuration("log.log_rotate_date")), // 日志切割时间间隔
	)
	if err != nil {
		logrus.Panic("Config local file system logger errcode.", errors.WithStack(err))
	}
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		//logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &logrus.TextFormatter{DisableColors: true, TimestampFormat: "2006-01-02 15:04:05.000"})
	logrus.AddHook(lfHook)
	logrus.AddHook(filename.NewHook()) // 记录日志的文件和行号

	gin.DefaultErrorWriter = writer // panic错误信息也记录到日志中
}

func (c Config) watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.Infof("Config file changed: %s", e.Name)
	})
}

func (c Config) initConfig() error {
	if c.Name != "" {
		viper.SetConfigFile(c.Name)
	} else {
		viper.AddConfigPath("conf")
		viper.SetConfigName("config")
	}
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("mog")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
