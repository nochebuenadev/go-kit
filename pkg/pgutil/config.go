package pgutil

import (
	"fmt"
	"net/url"
)

// DatabaseConfig holds the configuration parameters for connecting to a PostgreSQL database.
type DatabaseConfig struct {
	// Host is the database server host.
	Host string `env:"HOST,required,notEmpty"`
	// Port is the database server port.
	Port int `env:"PORT" envDefault:"5432"`
	// User is the database user.
	User string `env:"USER,required,notEmpty"`
	// Password is the database user password.
	Password string `env:"PASSWORD,required,notEmpty"`
	// Name is the database name.
	Name string `env:"NAME,required,notEmpty"`
	// SSLMode is the SSL mode for the connection (e.g., disable, require).
	SSLMode string `env:"SSL_MODE" envDefault:"disable"`
	// Timezone is the timezone for the database session.
	Timezone string `env:"TIMEZONE" envDefault:"UTC"`
	// MaxConns is the maximum number of connections in the pool.
	MaxConns int `env:"MAX_CONNS" envDefault:"5"`
	// MinConns is the minimum number of connections in the pool.
	MinConns int `env:"MIN_CONNS" envDefault:"2"`
	// MaxConnLifetime is the maximum amount of time a connection may be reused.
	MaxConnLifetime string `env:"MAX_CONN_LIFETIME" envDefault:"1h"`
	// MaxConnIdleTime is the maximum amount of time a connection may be idle.
	MaxConnIdleTime string `env:"MAX_CONN_IDLE_TIME" envDefault:"30m"`
	// HealthCheckPeriod is the period between health checks of connections.
	HealthCheckPeriod string `env:"HEALTH_CHECK_PERIOD" envDefault:"1m"`
}

// GetConnectionString constructs a PostgreSQL connection string from the configuration.
func (c *DatabaseConfig) GetConnectionString() string {
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
