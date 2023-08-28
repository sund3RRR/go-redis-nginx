package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	ServerPort    int    `yaml:"serverPort"`
	RedisHost     string `yaml:"redisHost"`
	RedisPort     int    `yaml:"redisPort"`
	RedisUser     string `yaml:"redisUser"`
	RedisPassword string `yaml:"redisPassword"`
}

func LoadConfig(filename string) (AppConfig, error) {
	var config AppConfig

	configFile, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}

	if err = yaml.Unmarshal(configFile, &config); err != nil {
		return config, err
	}

	return config, nil
}
