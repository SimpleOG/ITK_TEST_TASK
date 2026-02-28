package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"tryingMicro/OrderAccepter/internal/api/controllers"
	"tryingMicro/OrderAccepter/util/config"
)

type Server interface {
	Run(config config.Config) error
	Shutdown(ctx context.Context) error
}

type server struct {
	httpServer  *http.Server
	controllers *controllers.Controllers
	router      *gin.Engine
}

func NewServer(engine *gin.Engine, ctrl *controllers.Controllers) Server {
	return &server{
		httpServer:  nil,
		controllers: ctrl,
		router:      engine,
	}
}

func (s *server) Run(config config.Config) error {
	s.setupRoutes()
	s.httpServer = &http.Server{
		Addr:    config.ServerAddr,
		Handler: s.router,
	}
	if err := s.httpServer.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *server) setupRoutes() {
	api := s.router.Group("/api/v1")
	{
		wallets := api.Group("/wallets")
		{
			wallets.POST("/:walletId/operation", s.controllers.Wallet.ProcessOperation)
			wallets.GET("/:walletId", s.controllers.Wallet.GetBalance)
		}
	}
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
