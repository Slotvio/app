package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aleksandrboro/slotvio/internal/config"
)

var (
	errEmptyHandler = errors.New("HTTP handler is required")
	errEmptyHost    = errors.New("host is required")
	errEmptyPort    = errors.New("port is required")
)

type Server struct {
	log  *slog.Logger
	addr string
	srv  *http.Server
}

func NewServer(log *slog.Logger, cfg config.Config, handler http.Handler) (*Server, error) {
	// if handler == nil {
	// 	return nil, errEmptyHandler
	// }

	if cfg.Server.Host == "" {
		return nil, errEmptyHost
	}

	if cfg.Server.Port == "" {
		return nil, errEmptyPort
	}

	addr := strings.TrimSpace(fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port))

	return &Server{
		log:  log,
		addr: addr,
		srv: &http.Server{
			Addr:         addr,
			Handler:      handler,
			WriteTimeout: cfg.Server.WriteTimeout,
			ReadTimeout:  cfg.Server.ReadTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
	}, nil
}

func (s *Server) Start() error {
	s.log.Info("starting server", slog.String("addr", s.addr))

	if err := s.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("error starting server on %s: %w", s.addr, err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	s.log.Info("HTTP server stopped")

	return nil
}

func (s *Server) Close() error {
	return s.srv.Close()
}
