package initialize

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database parameter settings
const (
	dbTablePrefix = "orderx_"

	sqlBatchSize    = 1000
	maxIdleConns    = 10
	maxOpenConns    = 100
	connMaxLifetime = time.Hour
)

// Mysql mysql struct
type Mysql struct {
	DB       *gorm.DB
	host     string
	port     int
	user     string
	passwd   string
	database string
}

// NewMysql 初始化MySQL连接
func NewMysql() *gorm.DB {
	mysql := &Mysql{
		host:     viper.GetString("mysql.host"),
		port:     viper.GetInt("mysql.port"),
		user:     viper.GetString("mysql.user"),
		passwd:   viper.GetString("mysql.password"),
		database: viper.GetString("mysql.database"),
	}
	return mysql.connect().pool().DB
}

func (m *Mysql) connect() *Mysql {
	mysqlConfig := mysql.Config{
		DSN:                       m.dsn(), // DSN data source name
		DefaultStringSize:         256,     // string 类型字段的默认长度
		DisableDatetimePrecision:  true,    // 禁用 datetime 精度
		DontSupportRenameIndex:    true,    // 重命名索引时采用删除并新建的方式
		DontSupportRenameColumn:   true,    // 用 `change` 重命名列
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}

	// 设置日志级别
	logLevel := logger.Warn
	if viper.GetString("system.env") != "production" {
		logLevel = logger.Info // 非正式环境显示sql
	}

	// 创建自定义logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Init mysql failed, err: %v", err)
	}
	log.Println("Connected to MySQL!")

	m.DB = db
	return m
}

func (m *Mysql) dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.user, m.passwd, m.host, m.port, m.database)
}

func (m *Mysql) pool() *Mysql {
	sqlDB, err := m.DB.DB()
	if err != nil {
		log.Fatalf("get sql.DB failed: %v", err)
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	return m
}
