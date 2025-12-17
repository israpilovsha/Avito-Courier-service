package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	OrderServiceHost string
	Postgres         PostgresConfig
	Delivery         DeliveryConfig
	Kafka            KafkaConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type DeliveryConfig struct {
	TickerInterval time.Duration
}

type KafkaConfig struct {
	Enabled bool
	Brokers []string
	Topic   string
	GroupID string
}

func MustLoad() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	orderServiceHost := os.Getenv("ORDER_SERVICE_HOST")

	tickerRaw := os.Getenv("DELIVERY_TICKER_INTERVAL")
	if tickerRaw == "" {
		tickerRaw = "10s"
	}

	tickerInterval, err := time.ParseDuration(tickerRaw)
	if err != nil {
		panic("invalid DELIVERY_TICKER_INTERVAL: " + err.Error())
	}

	pg := PostgresConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
	}
	kafkaEnabled := os.Getenv("KAFKA_ENABLED") == "true"

	brokersRaw := os.Getenv("KAFKA_BROKERS")
	var brokers []string
	if brokersRaw != "" {
		brokers = strings.Split(brokersRaw, ",")
	}

	kafka := KafkaConfig{
		Enabled: kafkaEnabled,
		Brokers: brokers,
		Topic:   os.Getenv("KAFKA_TOPIC"),
		GroupID: os.Getenv("KAFKA_GROUP_ID"),
	}

	flag.StringVar(&port, "port", port, "Server port")
	flag.Parse()

	if port == "" {
		port = "8080"
	}

	if orderServiceHost == "" {
		panic("ORDER_SERVICE_HOST is required")
	}
	if kafka.Enabled {
		if len(kafka.Brokers) == 0 {
			panic("KAFKA_BROKERS is required when KAFKA_ENABLED=true")
		}
		if kafka.Topic == "" {
			panic("KAFKA_TOPIC is required when KAFKA_ENABLED=true")
		}
		if kafka.GroupID == "" {
			panic("KAFKA_GROUP_ID is required when KAFKA_ENABLED=true")
		}
	}

	return &Config{
		Port:             port,
		OrderServiceHost: orderServiceHost,
		Postgres:         pg,
		Delivery:         DeliveryConfig{TickerInterval: tickerInterval},
		Kafka:            kafka,
	}
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.DBName,
	)
}
