package configs

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type AppConfig struct {
	GRPCPort string   `env:"grpc_port" envDefault:":9090"`
	HTTPPort string   `env:"http_port" envDefault:":8080"`
	Domains  []string `env:"domains" envDefault:"phone_cellphones,phone_thegioididong"` // gold,diamond
	Workers  int      `env:"workers" envDefault:"100"`
}
type KafkaProducerConfig struct {
	Brokers string   `env:"producer_broker" envDefault:"localhost:29092"`
	Topic   []string `env:"producer_topic" envDefault:"normal,priority"`
}

type KafkaConsumerConfig struct {
	Brokers []string `env:"consumer_broker" envDefault:"localhost:29092"`
	Topic   []string `env:"consumer_topic" envDefault:"normal,priority"`
	GroupID string   `env:"consumer_group_id" envDefault:"crawler-local"`
}

type DatabaseConfig struct {
	Host     string `env:"db_host" envDefault:"localhost"`
	Port     int    `env:"db_port" envDefault:"5432"`
	User     string `env:"db_user" envDefault:"root"`
	Password string `env:"db_password" envDefault:"root"`
	DBName   string `env:"db_name" envDefault:"postgres"`
	SSLMode  string `env:"db_ssl_mode" envDefault:"disable"`
	Schema   string `env:"SCHEMA" envDefault:"public"`
}

type SchedulerService struct {
	Host    string        `env:"db_host" envDefault:"localhost:8080`
	Timeout time.Duration `env:"timeout" envDefault:"5s"`
}

type Redis struct {
	Addr     string `env:"redis_addr" envDefault:"localhost:6379"`
	Password string `env:"redis_password" envDefault:""`
	DB       int    `env:"redis_db" envDefault:"0"`
}

type Telegram struct {
	Enable      bool   `env:"telegram_enable" envDefault:"false"`
	APIKey      string `env:"telegram_api_key" envDefault:""`
	ChatId      int64  `env:"telegram_chat_id" envDefault:""`
	ChannelName string `env:"telegram_channel_name" envDefault:""`
}

type Config struct {
	AppConfig           AppConfig
	KafkaProducerConfig KafkaProducerConfig
	KafkaConsumerConfig KafkaConsumerConfig
	DatabaseConfig      DatabaseConfig
	Telegram            Telegram
	Redis               Redis
	SchedulerService    SchedulerService
}

func LoadConfig() *Config {
	var dbConfig Config
	if err := env.Parse(&dbConfig); err != nil {
		panic(err)
	}
	return &dbConfig
}
