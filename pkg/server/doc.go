/*
Package server provides a wrapper around the Echo web framework for building HTTP servers.

It simplifies the creation of a singleton HTTP server with pre-configured middlewares,
lifecycle management (initialization, startup, and graceful shutdown), and integrated
logging and error handling.

Features:
- Singleton HttpServer implementation based on Echo.
- Pre-configured CORS, Recovery, RequestID, and Logging middlewares.
- Custom error handling integrated with the mw package.
- Graceful shutdown support.
- Flexible route registration and grouping.

Example usage:

	cfg := &server.ServerConfig{Port: 8080, AllowedOrigins: []string{"*"}}
	srv := server.GetEchoServer(cfg, logger)

	if err := srv.OnInit(); err != nil {
		logger.Fatal("failed to init server", err)
	}

	srv.Registry(func(e *echo.Echo) {
		e.GET("/health", func(c echo.Context) error {
			return c.String(http.StatusOK, "UP")
		})
	})

	if err := srv.OnStart(); err != nil {
		logger.Fatal("failed to start server", err)
	}
*/
package server
