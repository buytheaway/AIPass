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
	"github.com/aipass/aipass/internal/logger"
	"github.com/aipass/aipass/internal/metrics"
	"github.com/aipass/aipass/internal/repository"
	"github.com/aipass/aipass/internal/service"
	accesshttp "github.com/aipass/aipass/internal/transport/http/accessapi"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load("access-api")
	log := logger.New(cfg.App.Env)
	defer log.Sync()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Warn("database unavailable; starting access-api in health-only stub mode", zap.Error(err))
		registry := metrics.NewRegistry(cfg.App.Name)
		e := echo.New()
		registerHealthOnly(e, cfg, registry)
		go startHTTP(log, e, cfg.HTTP.Addr)
		waitForShutdown(log, e)
		return
	}
	defer database.Close()

	registry := metrics.NewRegistry(cfg.App.Name)
	repos := repository.NewStore(database)
	services := service.NewContainer(cfg, repos, log)

	e := echo.New()
	accesshttp.Register(e, cfg, services, registry, log)

	go startHTTP(log, e, cfg.HTTP.Addr)

	waitForShutdown(log, e)
}

func registerHealthOnly(e *echo.Echo, cfg config.Config, registry *metrics.Registry) {
	e.HideBanner = true
	e.Use(registry.Middleware())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "degraded",
			"service": cfg.App.Name,
			"reason":  "database_unavailable",
		})
	})
	e.GET("/ready", func(c echo.Context) error {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status":  "not_ready",
			"service": cfg.App.Name,
			"reason":  "database_unavailable",
		})
	})
	e.GET("/metrics", registry.Handler())
}

func startHTTP(log *zap.Logger, e *echo.Echo, addr string) {
	if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("http server failed", zap.Error(err))
	}
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
