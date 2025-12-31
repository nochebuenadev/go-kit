package vkutil

// Config defines the configuration for the Valkey client.
type Config struct {
	// Addrs is a list of Valkey addresses to connect to.
	Addrs []string `env:"VK_ADDRS,required" envSeparator:","`
	// Password is the password for authentication.
	Password string `env:"VK_PASSWORD"`
	// SelectDB is the database number to select after connecting.
	SelectDB int `env:"VK_DB" envDefault:"0"`
	// CacheSizeEachConn is the size in MB for the client-side cache per connection.
	CacheSizeEachConn int `env:"VK_CLIENT_CACHE_MB" envDefault:"0"`
}
