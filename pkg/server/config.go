package server

// Config defines the configuration for the HTTP server.
type Config struct {
	// Host is the host address to bind the server to.
	Host string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	// Port is the port number to listen on.
	Port int `env:"SERVER_PORT" envDefault:"1323"`
	// AllowedOrigins is a list of origins for CORS configuration.
	AllowedOrigins []string `env:"SERVER_ALLOWED_ORIGINS,required" envSeparator:","`
}
