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
	"github.com/nochebuenadev/go-kit/pkg/launcher"
	"github.com/nochebuenadev/go-kit/pkg/logz"
	"github.com/nochebuenadev/go-kit/pkg/mw"
)

type (
	// RouterProvider defines the interface for registering routes and groups.
	RouterProvider interface {
		// Registry allows registering routes directly on the Echo instance.
		Registry(fn func(e *echo.Echo))
		// Group creates a new route group with the given prefix.
		Group(prefix string) *echo.Group
	}

	// HttpServerComponent extends RouterProvider with lifecycle management methods.
	HttpServerComponent interface {
		launcher.Component
		RouterProvider
	}

	// echoServer is the concrete implementation of HttpServerComponent using Labstack Echo.
	echoServer struct {
		// instance is the underlying Echo web framework instance.
		instance *echo.Echo
		// logger is used for tracking server events and requests.
		logger logz.Logger
		// cfg is the server configuration.
		cfg *Config
	}
)

var (
	// serverInstance is the singleton HTTP server component.
	serverInstance HttpServerComponent
	// serverOnce ensures the server is initialized only once.
	serverOnce sync.Once
)

// GetEchoServer returns the singleton instance of the HTTP server.
func GetEchoServer(cfg *Config, logger logz.Logger) HttpServerComponent {
	serverOnce.Do(func() {
		serverInstance = &echoServer{
			instance: echo.New(),
			logger:   logger,
			cfg:      cfg,
		}
	})

	return serverInstance
}

// OnInit implements the launcher.Component interface to initialize the Echo instance with standard middlewares and error handling.
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

	s.instance.Use(mw.WithRequestID())

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

// OnStart implements the launcher.Component interface to start the Echo server in a separate goroutine.
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

// OnStop implements the launcher.Component interface to gracefully shut down the Echo server with a timeout.
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
