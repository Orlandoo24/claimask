package initialize

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
)

// Server server
type Server struct {
	httpServer *http.Server
	Engine     *gin.Engine

	mysqlDB *gorm.DB
}

var server *Server

// NewServer 创建服务器
func NewServer(configPath string) *Server {
	// 加载配置文件
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	db, err := InitDB()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	e := gin.Default()
	server = &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", viper.GetString("server.port")),
			Handler: e,
		},
		Engine:  e,
		mysqlDB: db,
	}

	return server
}

// AddServer add server
func (s *Server) AddServer(serverFunc func(e *gin.Engine)) *Server {
	funcName := runtime.FuncForPC(reflect.ValueOf(serverFunc).Pointer()).Name()
	log.Printf("[服务注册] 服务名称: %s\n", funcName)
	serverFunc(s.Engine)
	return s
}

// AsyncStart async start
func (s *Server) AsyncStart() {
	log.Printf("[服务启动] 服务地址: %s\n", s.httpServer.Addr)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("[服务启动] 服务异常: %v\n", err)
		}
	}()
}

// Stop stop
func (s *Server) Stop() {
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("[服务关闭] 关闭服务")
	if err := s.httpServer.Shutdown(c); err != nil {
		log.Fatalf("[服务关闭] 关闭服务异常: %v\n", err)
	}
}

// GetMysqlInstance get mysql instance
func GetMysqlInstance() *gorm.DB {
	return server.mysqlDB
}

// InitDB 初始化数据库连接并迁移表结构
func InitDB() (*gorm.DB, error) {
	dbUser := viper.GetString("mysql.user")
	dbPassword := viper.GetString("mysql.password")
	dbHost := viper.GetString("mysql.host")
	dbPort := viper.GetInt("mysql.port")
	dbName := viper.GetString("mysql.database")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 设置连接池
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)

	// 开启日志
	db.LogMode(true)

	return db, nil
}
