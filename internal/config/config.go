package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App         AppConfig
	HTTP        HTTPConfig
	DatabaseURL string
	Auth        AuthConfig
	Redis       RedisConfig
	Kafka       KafkaConfig
	MinIO       MinIOConfig
	Telegram    TelegramConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Addr string
}

type AuthConfig struct {
	PrivateKeyPEM  string
	PublicKeyPEM   string
	PrivateKeyPath string
	PublicKeyPath  string
	AccessTokenTTL time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers           []string
	AccessEventsTopic string
	PaymentsTopic     string
	ConsumerGroup     string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	UserPhotos      string
	CheckinPhotos   string
	PaymentReceipts string
	Reports         string
}

type TelegramConfig struct {
	BotToken string
	ChatID   string
}

func Load(serviceName string) Config {
	defaultAddr := ":8080"
	if serviceName == "scanner-gateway" {
		defaultAddr = ":8081"
	}
	if serviceName == "notification-report-service" {
		defaultAddr = ":8082"
	}

	return Config{
		App:         AppConfig{Name: env("APP_NAME", serviceName), Env: env("APP_ENV", "local")},
		HTTP:        HTTPConfig{Addr: env("HTTP_ADDR", defaultAddr)},
		DatabaseURL: env("DATABASE_URL", "postgres://aipass:aipass@localhost:5432/aipass?sslmode=disable"),
		Auth: AuthConfig{
			PrivateKeyPEM:  env("JWT_PRIVATE_KEY_PEM", ""),
			PublicKeyPEM:   env("JWT_PUBLIC_KEY_PEM", ""),
			PrivateKeyPath: env("JWT_PRIVATE_KEY_PATH", "deployments/docker/keys/private.pem"),
			PublicKeyPath:  env("JWT_PUBLIC_KEY_PATH", "deployments/docker/keys/public.pem"),
			AccessTokenTTL: time.Duration(envInt("JWT_ACCESS_TTL_MINUTES", 60)) * time.Minute,
		},
		Redis: RedisConfig{Addr: env("REDIS_ADDR", "localhost:6379"), Password: env("REDIS_PASSWORD", ""), DB: envInt("REDIS_DB", 0)},
		Kafka: KafkaConfig{
			Brokers:           envList("KAFKA_BROKERS", "localhost:9092"),
			AccessEventsTopic: env("KAFKA_ACCESS_EVENTS_TOPIC", "access.events.v1"),
			PaymentsTopic:     env("KAFKA_PAYMENTS_TOPIC", "payments.events.v1"),
			ConsumerGroup:     env("KAFKA_CONSUMER_GROUP", "notification-report-service"),
		},
		MinIO: MinIOConfig{
			Endpoint:        env("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:       env("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:       env("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          envBool("MINIO_USE_SSL", false),
			UserPhotos:      env("MINIO_BUCKET_USER_PHOTOS", "aipass-user-photos"),
			CheckinPhotos:   env("MINIO_BUCKET_CHECKIN_PHOTOS", "aipass-checkin-photos"),
			PaymentReceipts: env("MINIO_BUCKET_PAYMENT_RECEIPTS", "aipass-payment-receipts"),
			Reports:         env("MINIO_BUCKET_REPORTS", "aipass-reports"),
		},
		Telegram: TelegramConfig{BotToken: env("TELEGRAM_BOT_TOKEN", ""), ChatID: env("TELEGRAM_CHAT_ID", "")},
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(env(key, ""))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}

func envList(key, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
