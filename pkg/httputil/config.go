package httputil

import (
	"time"
)

// Config holds the configuration for the resilient HTTP client.
type Config struct {
	// Timeout is the total timeout for the request, including retries.
	Timeout time.Duration `env:"HTTP_TIMEOUT" envDefault:"30s"`
	// DialTimeout is the maximum time to wait for a connection to be established.
	DialTimeout time.Duration `env:"HTTP_DIAL_TIMEOUT" envDefault:"5s"`
	// MaxRetries is the maximum number of retries for failed requests (5xx or network errors).
	MaxRetries uint `env:"HTTP_MAX_RETRIES" envDefault:"3"`
	// RetryDelay is the initial delay for exponential backoff.
	RetryDelay time.Duration `env:"HTTP_RETRY_DELAY" envDefault:"1s"`
	// CBThreshold is the number of consecutive failures before the circuit breaker opens.
	CBThreshold uint32 `env:"HTTP_CB_THRESHOLD" envDefault:"10"`
	// CBTimeout is the duration the circuit breaker stays open before transitioning to half-open.
	CBTimeout time.Duration `env:"HTTP_CB_TIMEOUT" envDefault:"1m"`
}

// DefaultConfig returns a default configuration for the HTTP client.
func DefaultConfig() *Config {
	return &Config{
		Timeout:     30 * time.Second,
		DialTimeout: 5 * time.Second,
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		CBThreshold: 10,
		CBTimeout:   1 * time.Minute,
	}
}
