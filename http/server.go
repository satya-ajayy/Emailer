package http

import (
	// Go Internal Packages
	"context"
	"net/http"
	"time"

	// Local Packages
	errors "emailer/errors"
	handlers "emailer/http/handlers"
	apxresp "emailer/http/response"
	kafka "emailer/kafka"
	health "emailer/services/health"

	// External Packages
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"moul.io/chizap"
)

// Server struct follows the alphabet order
type Server struct {
	consumer *kafka.Consumer
	health   *health.HealthCheckService
	logger   *zap.Logger
	prefix   string
	emails   *handlers.EmailHandler
}

func NewServer(prefix string, logger *zap.Logger, consumer *kafka.Consumer, healthCheck *health.HealthCheckService, emails *handlers.EmailHandler) *Server {
	return &Server{
		consumer: consumer,
		logger:   logger,
		prefix:   prefix,
		emails:   emails,
		health:   healthCheck,
	}
}

func (s *Server) Listen(ctx context.Context, addr string) error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(chizap.New(s.logger, &chizap.Opts{
		WithReferer:   false,
		WithUserAgent: false,
	}))
	r.Use(middleware.Recoverer)

	r.Route(s.prefix, func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/health", s.HealthCheckHandler)
			r.Post("/send", s.ToHTTPHandlerFunc(s.emails.Send))
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

// ToHTTPHandlerFunc converts a handler function to an http.HandlerFunc.
// This wrapper function is used to handle errors and respond to the client
func (s *Server) ToHTTPHandlerFunc(handler func(w http.ResponseWriter, r *http.Request) (any, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, status, err := handler(w, r)
		if err != nil {
			switch err := err.(type) {
			case *errors.Error:
				apxresp.RespondError(w, err)
			default:
				s.logger.Error("internal error", zap.Error(err))
				apxresp.RespondMessage(w, http.StatusInternalServerError, "internal error")
			}
			return
		}
		if response != nil {
			apxresp.RespondJSON(w, status, response)
		}
		if status >= 100 && status < 600 {
			w.WriteHeader(status)
		}
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
