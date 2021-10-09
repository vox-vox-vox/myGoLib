package global

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

const (
	transactionKindNone transactionKind = iota
	transactionKindNormal
	transactionKindInternal
	defaultMaxOpenConns = 1000
)

var GpgDb *DB

type transactionKind int
type DB struct {
	*gorm.DB
	driver          string
	config          string
	transactionKind transactionKind
}

func InitStore() {
	// init once only
	if GpgDb != nil {
		return
	}
	LoadConfig()
	GpgDb = NewPgDBWithConfig("postgresConfig")
}

func NewPgDBWithConfig(pgConfigName string) *DB {
	user := viper.GetString(pgConfigName + ".user")
	passwd := viper.GetString(pgConfigName + ".passwd")
	host := viper.GetString(pgConfigName + ".host")
	port := viper.GetInt(pgConfigName + ".port")
	dbName := viper.GetString(pgConfigName + ".dbname")
	pgConnectStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", host, port, user, dbName, passwd)
	pgDB := NewDB("postgres", pgConnectStr)
	maxOpenConns := viper.GetInt(pgConfigName + ".maxOpenConns")
	if maxOpenConns == 0 {
		maxOpenConns = defaultMaxOpenConns
	}

	sqlDB, _ := pgDB.DB.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(maxOpenConns)

	return pgDB
}


func NewDB(driver, connectStr string) *DB {
	// gorm 自带logger，可以按照如下的方法配置：
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold: time.Second,   // 慢 SQL 阈值
			LogLevel:      logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,   // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful: true,
		},
	)
	db, err := gorm.Open(postgres.Open(connectStr),
		// gorm2.0 默认写操作开启事务，为了测验实验效果，将其关闭
		&gorm.Config{
			SkipDefaultTransaction: true,
			Logger: newLogger,
		})
	if err!=nil{
		panic(err)
	}
	return &DB{
		DB:     db,
		driver: driver,
		config: connectStr,
	}
}

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./sql")
	err := viper.ReadInConfig()
	if err != nil {             // Handle errors reading the config file
		log.Fatal("fail to load config file", err)
	}
}

