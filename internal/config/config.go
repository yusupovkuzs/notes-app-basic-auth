package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string           `yaml:"env"`
	Postgres   PostgresConfig   `yaml:"postgres"`
	HttpServer HttpServerConfig `yaml:"http_server"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	DBName   string `yaml:"db_name"`
	Password string `env:"POSTGRES_PASSWORD"`
}

type HttpServerConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file")
	}
	configPath := os.Getenv("CONFIG_PATH")
	if err := os.Setenv("CONFIG_PATH", configPath); err != nil {
		log.Fatal("CONFIG_PATH is not set")
		return nil
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	return &cfg
}
