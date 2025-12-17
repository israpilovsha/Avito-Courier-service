package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	OrderServiceHost string
	Postgres         PostgresConfig
	Delivery         DeliveryConfig
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

	flag.StringVar(&port, "port", port, "Server port")
	flag.Parse()

	if port == "" {
		port = "8080"
	}

	if orderServiceHost == "" {
		panic("ORDER_SERVICE_HOST is required")
	}

	return &Config{
		Port:             port,
		OrderServiceHost: os.Getenv("ORDER_SERVICE_HOST"),
		Postgres:         pg,
		Delivery:         DeliveryConfig{TickerInterval: tickerInterval},
	}
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.DBName,
	)
}
