package conf

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type RPCConfig struct {
	IP       string
	Port     int
	User     string
	Password string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ServerConfig struct {
	Port string
}

type WalletGroup struct {
	Group          int
	Receive        string
	ReceivePrivate string
}

func LoadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./conf")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		zap.L().Fatal("Failed to read config file", zap.Error(err))
	}
}
