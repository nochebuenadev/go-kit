package mysqlutil

import (
	"fmt"
)

// Config holds the configuration parameters for connecting to a MySQL database.
type Config struct {
	// Host is the database server host.
	Host string `env:"MYSQL_HOST,required"`
	// Port is the database server port.
	Port int `env:"MYSQL_PORT" envDefault:"3306"`
	// User is the database user.
	User string `env:"MYSQL_USER,required"`
	// Password is the database user password.
	Password string `env:"MYSQL_PASSWORD,required"`
	// Name is the database name.
	Name string `env:"MYSQL_NAME,required"`
	// MaxConns is the maximum number of connections in the pool.
	MaxConns int `env:"MYSQL_MAX_CONNS" envDefault:"5"`
	// MinConns is the minimum number of connections in the pool.
	MinConns int `env:"MYSQL_MIN_CONNS" envDefault:"2"`
	// MaxConnLifetime is the maximum amount of time a connection may be reused.
	MaxConnLifetime string `env:"MYSQL_MAX_CONN_LIFETIME" envDefault:"1h"`
	// MaxConnIdleTime is the maximum amount of time a connection may be idle.
	MaxConnIdleTime string `env:"MYSQL_MAX_CONN_IDLE_TIME" envDefault:"30m"`
}

// GetConnectionString constructs a MySQL DSN (Data Source Name).
// Format: username:password@tcp(host:port)/dbname?parseTime=true
func (c *Config) GetConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.Name)
}
