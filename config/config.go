package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	CronSpec    string
	WithinHours int32

	SubscriptionAddr string
	UserServiceAddr  string
	RabbitMQURL      string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port: env("PORT", "8080"),

		CronSpec:    env("CRON_SPEC", "@every 1h"),
		WithinHours: envInt32("WITHIN_HOURS", 24),

		SubscriptionAddr: env("SUBSCRIPTION_SERVICE_ADDR", "localhost:50051"),
		UserServiceAddr:  env("USER_SERVICE_ADDR", "localhost:50052"),
		RabbitMQURL:      env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt32(k string, def int32) int32 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return def
	}
	return int32(n)
}
