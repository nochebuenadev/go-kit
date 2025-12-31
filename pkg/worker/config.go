package worker

// Config defines the configuration for the background worker pool.
type Config struct {
	// PoolSize is the number of concurrent workers processing tasks.
	PoolSize int `env:"WORKER_POOL_SIZE" envDefault:"5"`
	// BufferSize is the capacity of the task queue channel.
	BufferSize int `env:"WORKER_BUFFER_SIZE" envDefault:"100"`
}
