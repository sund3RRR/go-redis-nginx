package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	ServerPort    int    `yaml:"server_port"`
	RedisHost     string `yaml:"redis_host"`
	RedisPort     int    `yaml:"redis_port"`
	RedisUser     string `yaml:"redis_user"`
	RedisPassword string `yaml:"redis_password"`
	UseRedisTLS   bool   `yaml:"use_redis_tls"`
	ZapConfig     zap.Config
}

func NewConfig(filename string) (*AppConfig, error) {
	var config AppConfig

	configFile, err := os.ReadFile(filename)
	if err != nil {
		return &config, err
	}

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	config.ZapConfig = zapConfig

	err = yaml.Unmarshal(configFile, &config)
	return &config, err
}
