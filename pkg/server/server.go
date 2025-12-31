package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nochebuenadev/go-kit/pkg/logz"
	"github.com/nochebuenadev/go-kit/pkg/mw"
)

type (
	// Router defines the interface for registering routes and groups.
	Router interface {
		// Registry allows registering routes directly on the Echo instance.
		Registry(fn func(e *echo.Echo))
		// Group creates a new route group with the given prefix.
		Group(prefix string) *echo.Group
	}

	// HttpServer extends Router with lifecycle management methods.
	HttpServer interface {
		Router
		// OnInit initializes the server with middlewares and configurations.
		OnInit() error
		// OnStart starts the HTTP server in a background goroutine.
		OnStart() error
		// OnStop performs a graceful shutdown of the HTTP server.
		OnStop() error
	}

	// echoServer is the concrete implementation of HttpServer using Labstack Echo.
	echoServer struct {
		instance *echo.Echo
		logger   logz.Logger
		cfg      *Config
	}
)

var (
	serverInstance HttpServer
	serverOnce     sync.Once
)

// GetEchoServer returns the singleton instance of the HTTP server.
func GetEchoServer(cfg *Config, logger logz.Logger) HttpServer {
	serverOnce.Do(func() {
		serverInstance = &echoServer{
			instance: echo.New(),
			logger:   logger,
			cfg:      cfg,
		}
	})

	return serverInstance
}

// OnInit initializes the Echo instance with standard middlewares and error handling.
func (s *echoServer) OnInit() error {
	s.instance.HideBanner = true
	s.instance.HidePort = true

	s.instance.HTTPErrorHandler = mw.AppErrorHandler(s.logger)

	s.instance.Use(middleware.Recover())

	s.instance.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: s.cfg.AllowedOrigins,
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	s.instance.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.New().String()
		},
	}))

	s.instance.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogMethod:    true,
		LogURI:       true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				s.logger.With(
					"uri", v.URI,
					"method", v.Method,
					"status", v.Status,
					"request_id", v.RequestID,
				).Info("PETICIÓN")
			} else {
				s.logger.With(
					"uri", v.URI,
					"method", v.Method,
					"status", v.Status,
					"request_id", v.RequestID,
				).Error("PETICIÓN_FALLIDA", v.Error)
			}
			return nil
		},
	}))

	return nil
}

// OnStart starts the Echo server in a separate goroutine.
func (s *echoServer) OnStart() error {
	s.logger.Info("Iniciando servidor HTTP", "port", s.cfg.Port)

	go func() {
		addr := fmt.Sprintf(":%d", s.cfg.Port)
		if err := s.instance.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Error fatal en el servidor", err)
		}
	}()

	return nil
}

// OnStop gracefully shuts down the Echo server with a timeout.
func (s *echoServer) OnStop() error {
	s.logger.Info("Apagando servidor HTTP (Graceful Shutdown)...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.instance.Shutdown(ctx)
}

// Registry registers routes using the provided function.
func (s *echoServer) Registry(fn func(e *echo.Echo)) {
	fn(s.instance)
}

// Group returns a new Echo route group.
func (s *echoServer) Group(prefix string) *echo.Group {
	return s.instance.Group(prefix)
}
