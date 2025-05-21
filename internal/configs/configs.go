package configs

import "github.com/caarlos0/env/v6"

type AppConfig struct {
	GRPCPort string `env:"grpc_port" envDefault:":9090"`
	HTTPPort string `env:"http_port" envDefault:"8080"`
}
type KafkaProducerConfig struct {
	Broker string `env:"producer_broker" envDefault:"localhost:9092"`
	Topic  string `env:"producer_topic" envDefault:"crawler"`
}

type KafkaConsumerConfig struct {
	Broker string `env:"consumer_broker" envDefault:"localhost:9092"`
	Topic  string `env:"consumer_topic" envDefault:"crawler"`
}

type DatabaseConfig struct {
	Host     string `env:"db_host" envDefault:"localhost"`
	Port     int    `env:"db_port" envDefault:"5432"`
	User     string `env:"db_user" envDefault:"root"`
	Password string `env:"db_password" envDefault:"root"`
	DBName   string `env:"db_name" envDefault:"postgres"`
	SSLMode  string `env:"db_ssl_mode" envDefault:"disable"`
}

type Config struct {
	AppConfig           AppConfig
	KafkaProducerConfig KafkaProducerConfig
	KafkaConsumerConfig KafkaConsumerConfig
	DatabaseConfig      DatabaseConfig
}

func LoadConfig() *Config {
	var dbConfig Config
	if err := env.Parse(&dbConfig); err != nil {
		panic(err)
	}
	return &dbConfig
}
