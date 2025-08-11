package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
	"tz1/pkg/logging"
)

type Config struct {
	Listen struct {
		BindIp string `yaml:"bind_ip" env-default:""`
		Port   string `yaml:"port" env-default:"8080"`
	} `yaml:"listen"`
	Storage StorageConfig `yaml:"storage"`
}

type StorageConfig struct {
	Host     string `yaml:"host" env-default:"postgres"`
	Port     string `yaml:"port" env-default:"5432"`
	Database string `yaml:"database" env-default:"tz1"`
	Username string `yaml:"username" env-default:"local"`
	Password string `yaml:"password" env-default:"admin"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read application configuration")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}
