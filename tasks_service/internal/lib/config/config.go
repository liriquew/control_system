package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	TasksService      ServiceConfig         `yaml:"service_config" env-required:"true"`
	AuthClient        ClientConfig          `yaml:"auth_service" env-required:"true"`
	PredictionsClient ClientConfig          `yaml:"predictions_service" env-required:"true"`
	GroupsClient      ClientConfig          `yaml:"groups_service" env-required:"true"`
	GraphsClient      ClientConfig          `yaml:"graphs_service" env-required:"true"`
	Storage           StorageConfig         `yaml:"postgres" env-required:"true"`
	KafkaConfig       KafkaTasksTopicConfig `yaml:"kafka" env-required:"true"`
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

type KafkaTasksTopicConfig struct {
	ConnStr     string `yaml:"bootstrap_servers" env-required:"true"`
	Topic       string `yaml:"topic" env-required:"true"`
	DeleteTopic string `yaml:"topic_delete" env-required:"true"`
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
