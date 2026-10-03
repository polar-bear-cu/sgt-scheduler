package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	ReminderCron string
	RolloverCron string
	PublicURL    string

	SubscriptionAddr string
	UserServiceAddr  string
	RabbitMQURL      string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port: env("PORT", "8080"),

		ReminderCron: env("REMINDER_CRON", "0 9 * * *"),
		RolloverCron: env("ROLLOVER_CRON", "0 0 * * *"),
		PublicURL:    env("PUBLIC_URL", "http://localhost:8000"),

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
