package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
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

func NewServer(engine *gin.Engine) Server {
	return &server{
		httpServer: nil,
		router:     engine,
	}
}
func (s *server) Run(config config.Config) error {
	s.setupRoutes()
	s.httpServer = &http.Server{
		Addr:           config.ServerAddr,
		Handler:        s.router,
		ReadTimeout:    0,
		WriteTimeout:   0,
		IdleTimeout:    0,
		MaxHeaderBytes: 0,
	}
	if err := s.httpServer.ListenAndServe(); err != nil {
		return err
	}
	return nil

}
func (s *server) setupRoutes() {

}

func (s *server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
