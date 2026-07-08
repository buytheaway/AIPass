package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aipass/aipass/internal/config"
	"github.com/aipass/aipass/internal/db"
	"github.com/aipass/aipass/internal/kafka"
	"github.com/aipass/aipass/internal/logger"
	"github.com/aipass/aipass/internal/metrics"
	redisclient "github.com/aipass/aipass/internal/redis"
	"github.com/aipass/aipass/internal/repository"
	"github.com/aipass/aipass/internal/service"
	scannerhttp "github.com/aipass/aipass/internal/transport/http/scanner"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load("scanner-gateway")
	log := logger.New(cfg.App.Env)
	defer log.Sync()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("connect database", zap.Error(err))
	}
	defer database.Close()

	redisClient := redisclient.Connect(cfg.Redis)
	defer redisClient.Close()

	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.AccessEventsTopic)
	defer producer.Close()

	registry := metrics.NewRegistry(cfg.App.Name)
	repos := repository.NewStore(database)
	services := service.NewContainer(cfg, repos, log)
	services.Access.Redis = redisClient
	services.Access.Events = producer

	e := echo.New()
	scannerhttp.Register(e, cfg, services, registry, log)

	go func() {
		if err := e.Start(cfg.HTTP.Addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http server failed", zap.Error(err))
		}
	}()

	waitForShutdown(log, e)
}

func waitForShutdown(log *zap.Logger, e *echo.Echo) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Error("shutdown failed", zap.Error(err))
	}
}
