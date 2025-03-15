package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	PredictionsService ServiceConfig `yaml:"predictions_service" env-required:"true"`
	APIService         ServiceConfig `yaml:"api_service" env-required:"true"`
	Storage            StorageConfig `yaml:"postgres" env-required:"true"`
}

type ServiceConfig struct {
	Port int    `yaml:"port" env-required:"true"`
	Host string `yaml:"host" env-required:"true"`
}

type StorageConfig struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DBName   string `yaml:"db_name" env-required:"true"`
}

func MustLoad() AppConfig {
	path := fetchConfigPath()

	if path == "" {
		path = "./config/config.yaml"
	}

	return MustLoadPath(path)
}

func fetchConfigPath() string {
	return os.Getenv("CONF_PATH")
}

func MustLoadPath(path string) AppConfig {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config: file not exist")
	}

	var cfg AppConfig

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("error while reading config" + err.Error())
	}

	return cfg
}
