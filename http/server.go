package http

import (
	// Go Internal Packages
	"context"
	"net/http"
	"time"

	// Local Packages
	smiddleware "emailer/http/middlewares"
	apxresp "emailer/http/response"
	kafka "emailer/kafka"
	health "emailer/services/health"

	// External Packages
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
)

// Server struct follows the alphabet order
type Server struct {
	consumer *kafka.Consumer
	health   *health.HealthCheckService
	logger   *zap.Logger
	prefix   string
}

func NewServer(prefix string, logger *zap.Logger, consumer *kafka.Consumer, healthCheck *health.HealthCheckService) *Server {
	return &Server{
		consumer: consumer,
		logger:   logger,
		prefix:   prefix,
		health:   healthCheck,
	}
}

func (s *Server) Listen(ctx context.Context, addr string) error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(smiddleware.HTTPMiddleware(s.logger))
	r.Use(middleware.Recoverer)

	r.Route(s.prefix, func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/health", s.HealthCheckHandler)
		})
	})

	errch := make(chan error)
	server := &http.Server{Addr: addr, Handler: r}
	go func() {
		s.logger.Info("starting server", zap.String("addr", addr))
		errch <- server.ListenAndServe()
	}()

	select {
	case err := <-errch:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

// HealthCheckHandler returns the health status of the service
func (s *Server) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if ok := s.health.Health(r.Context()); !ok {
		apxresp.RespondMessage(w, http.StatusServiceUnavailable, "health check failed")
		return
	}
	apxresp.RespondMessage(w, http.StatusOK, "!!! We are RunninGoo !!!")
}
