package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	API               ServiceConfig `yaml:"api_gateway" env-required:"true"`
	TasksService      ClientConfig  `yaml:"tasks_service" env-required:"true"`
	AuthClient        ClientConfig  `yaml:"auth_service" env-required:"true"`
	PredictionsClient ClientConfig  `yaml:"predictions_service" env-required:"true"`
	GroupsClient      ClientConfig  `yaml:"groups_service" env-required:"true"`
	GraphsClient      ClientConfig  `yaml:"graphs_service" env-required:"true"`
}

type AppTestConfig struct {
	API ServiceConfig `yaml:"api_gateway" env-required:"true"`
}

type ServiceConfig struct {
	Port int    `yaml:"port" env-required:"true"`
	Host string `yaml:"host" env-required:"true"`
}

type ClientConfig struct {
	Port    string        `yaml:"port" env-required:"true"`
	Retries int           `yaml:"retries" env-required:"true"`
	Host    string        `yaml:"host" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-defauilt:"1s"`
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

func MustLoadPathTest(path string) AppTestConfig {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config: file not exist")
	}

	var cfg AppTestConfig

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("error while reading config" + err.Error())
	}

	return cfg
}
