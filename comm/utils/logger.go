package utils

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger 初始化日志
func InitLogger() {
	// 从配置获取日志级别
	logLevel := getLogLevel(viper.GetString("log.level"))

	// 创建基础配置
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	// 创建日志实例
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// 替换全局logger
	zap.ReplaceGlobals(logger)
}

// getLogLevel 获取日志级别
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
