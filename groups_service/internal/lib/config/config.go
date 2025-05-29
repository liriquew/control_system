package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	ServiceConfig ServiceConfig `yaml:"service_config" env-required:"true"`
	Storage       StorageConfig `yaml:"postgres" env-required:"true"`
}

type ServiceConfig struct {
	Port int `yaml:"port" env-required:"true"`
}

type ClientConfig struct {
	Port    string        `yaml:"port" env-required:"true"`
	Retries int           `yaml:"retries" env-required:"true"`
	Host    string        `yaml:"host" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-defauilt:"1s"`
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
