package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tryingMicro/OrderAccepter/internal/api/controllers"
	"tryingMicro/OrderAccepter/internal/api/server"
	"tryingMicro/OrderAccepter/internal/repository"
	"tryingMicro/OrderAccepter/internal/service"
	"tryingMicro/OrderAccepter/package/logger"
	"tryingMicro/OrderAccepter/util/config"
)

func main() {

	cfg, err := config.InitConfig(".")
	if err != nil {
		log.Panicf("failed to load config: %s", err)
	}

	logger, err := logger.NewLogger(zapcore.Level(cfg.LoggerLevel))
	if err != nil {
		log.Panicf("failed to build logger: %s", err)
	}

	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(cfg.DBURL())
	if err != nil {
		logger.Fatal("failed to parse db config", zap.Error(err))
	}
	poolConfig.MaxConns = cfg.DBMaxConns
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		logger.Fatal("database is not reachable", zap.Error(err))
	}
	logger.Info("connected to database")

	repo := repository.NewRepository(pool)
	services := service.NewServices(repo, logger)

	ctrls := controllers.NewControllers(services, logger)

	router := gin.Default()
	srv := server.NewServer(router, ctrls)

	errChan := make(chan error, 1)
	go func() {
		logger.Info("starting server", zap.String("addr", cfg.ServerAddr))
		if err = srv.Run(cfg); err != nil {
			errChan <- err
		}
	}()

	osChan := make(chan os.Signal, 1)
	signal.Notify(osChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case err = <-errChan:
		logger.Fatal(fmt.Sprintf("server error: %s", err))
	case sig := <-osChan:
		logger.Info("shutting down", zap.String("signal", sig.String()))
		if err = srv.Shutdown(ctx); err != nil {
			logger.Fatal("shutdown error", zap.Error(err))
		}
		logger.Info("server stopped")
	}
}
