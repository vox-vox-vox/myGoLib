package global

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

const (
	defaultMaxOpenConns = 1000
)

var GMySQLDb *gorm.DB

func InitStore() {
	// init once only
	if GMySQLDb != nil {
		return
	}
	LoadConfig()
	GMySQLDb = NewPgDBWithConfig("mysqlConfig")
}

func NewPgDBWithConfig(ConfigName string) *gorm.DB {
	user := viper.GetString(ConfigName + ".user")
	passwd := viper.GetString(ConfigName + ".passwd")
	host := viper.GetString(ConfigName + ".host")
	port := viper.GetInt(ConfigName + ".port")
	dbName := viper.GetString(ConfigName + ".dbname")
	ConnectStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", user, passwd, host, port, dbName)

	// gorm 自带logger，可以按照如下的方法配置：
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(mysql.Open(ConnectStr), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 newLogger,
	})

	if err != nil {
		panic(err)
	}

	maxOpenConns := viper.GetInt(ConfigName + ".maxOpenConns")
	if maxOpenConns == 0 {
		maxOpenConns = defaultMaxOpenConns
	}

	sqlDB, _ := db.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(maxOpenConns)

	return db
}

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./sql")
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Fatal("fail to load config file", err)
	}
}
