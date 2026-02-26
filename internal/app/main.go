package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tryingMicro/OrderAccepter/internal/api/controllers"
	"tryingMicro/OrderAccepter/internal/api/server"
	"tryingMicro/OrderAccepter/package/logger"
	"tryingMicro/OrderAccepter/util/config"
)

func main() {
	config, err := config.InitConfig(".")
	if err != nil {

	}
	logger, err := logger.NewLogger(zapcore.Level(config.LoggerLevel))
	if err != nil {
		log.Panicf("failed to build logger : %s", err)
	}
	ctx := context.Background()
	controllers.NewControllers(logger)
	router := gin.Default()

	// шатдаун
	errChan := make(chan error)
	server := server.NewServer(router)
	go func() {
		if err = server.Run(config); err != nil {
			errChan <- err
			return
		}
	}()
	osChan := make(chan os.Signal)
	signal.Notify(osChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case err = <-errChan:
		logger.Fatal(fmt.Sprintf("failed to start server : %s", err))
		os.Exit(1)
	case <-osChan:
		if err = server.Shutdown(ctx); err != nil {
			logger.Fatal(fmt.Sprintf("failed to shutdown server : %s", err))
		}
		os.Exit(1)
	}
}
