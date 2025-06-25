package configs

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	AppPort        string `env:"APP_PORT" envDefault:"8080"`
	DatabaseConfig DatabaseConfig
	AIConfig       AIConfig
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"root"`
	Password string `env:"DB_PASSWORD" envDefault:"root"`
	DBName   string `env:"DB_NAME" envDefault:"postgres"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

type AIConfig struct {
	Host string `env:"AI_HOST" envDefault:"http://127.0.0.1:8082/chat"`
}

func LoadConfig() *Config {
	var conf Config
	if err := env.Parse(&conf); err != nil {
		panic("cannot load env")
	}
	return &conf
}
