package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aipass/aipass/internal/config"
	"github.com/aipass/aipass/internal/kafka"
	"github.com/aipass/aipass/internal/logger"
	"github.com/aipass/aipass/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load("notification-report-service")
	log := logger.New(cfg.App.Env)
	defer log.Sync()

	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.AccessEventsTopic, cfg.Kafka.ConsumerGroup)
	defer consumer.Close()

	notifications := service.NewNotificationService(cfg, log)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("notification-report-service started")
	if err := notifications.ConsumeAccessEvents(ctx, consumer); err != nil {
		log.Error("notification service stopped", zap.Error(err))
	}
}
