package pgutil

import (
	"fmt"
	"net/url"
)

// Config holds the configuration parameters for connecting to a PostgreSQL database.
type Config struct {
	// Host is the database server host.
	Host string `env:"PG_HOST,required"`
	// Port is the database server port.
	Port int `env:"PG_PORT" envDefault:"5432"`
	// User is the database user.
	User string `env:"PG_USER,required"`
	// Password is the database user password.
	Password string `env:"PG_PASSWORD,required"`
	// Name is the database name.
	Name string `env:"PG_NAME,required"`
	// SSLMode is the SSL mode for the connection (e.g., disable, require).
	SSLMode string `env:"PG_SSL_MODE" envDefault:"disable"`
	// Timezone is the timezone for the database session.
	Timezone string `env:"PG_TIMEZONE" envDefault:"UTC"`
	// MaxConns is the maximum number of connections in the pool.
	MaxConns int `env:"PG_MAX_CONNS" envDefault:"5"`
	// MinConns is the minimum number of connections in the pool.
	MinConns int `env:"PG_MIN_CONNS" envDefault:"2"`
	// MaxConnLifetime is the maximum amount of time a connection may be reused.
	MaxConnLifetime string `env:"PG_MAX_CONN_LIFETIME" envDefault:"1h"`
	// MaxConnIdleTime is the maximum amount of time a connection may be idle.
	MaxConnIdleTime string `env:"PG_MAX_CONN_IDLE_TIME" envDefault:"30m"`
	// HealthCheckPeriod is the period between health checks of connections.
	HealthCheckPeriod string `env:"PG_HEALTH_CHECK_PERIOD" envDefault:"1m"`
}

// GetConnectionString constructs a PostgreSQL connection string from the configuration.
func (c *Config) GetConnectionString() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   "/" + c.Name,
	}

	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	q.Set("timezone", c.Timezone)

	u.RawQuery = q.Encode()

	return u.String()
}
