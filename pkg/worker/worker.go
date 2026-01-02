package worker

import (
	"context"
	"sync"

	"github.com/nochebuenadev/go-kit/pkg/launcher"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// Task represents a unit of work to be executed asynchronously.
	Task func(ctx context.Context) error

	// Provider defines the interface for dispatching tasks to the worker pool.
	Provider interface {
		// Dispatch adds a task to the queue. Returns true if the task was queued,
		// or false if the queue is full (backpressure).
		Dispatch(task Task) bool
	}

	// Component extends Provider with lifecycle management methods.
	Component interface {
		launcher.Component
		Provider
	}

	// workerComponent is the concrete implementation of the worker pool.
	workerComponent struct {
		// logger is used for tracking task execution and errors.
		logger logz.Logger
		// cfg is the worker pool configuration.
		cfg *Config
		// taskQueue is the channel used to dispatch tasks to workers.
		taskQueue chan Task
		// wg tracks the lifecycle of active worker goroutines.
		wg sync.WaitGroup
		// ctx is the background context for worker tasks.
		ctx context.Context
		// cancel is used to signal workers to stop.
		cancel context.CancelFunc
	}
)

var (
	// instance is the singleton worker component.
	instance Component
	// once ensures that the worker pool is initialized only once.
	once sync.Once
)

// GetWorker returns the singleton instance of the worker component.
func GetWorker(logger logz.Logger, cfg *Config) Component {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		instance = &workerComponent{
			logger:    logger,
			cfg:       cfg,
			taskQueue: make(chan Task, cfg.BufferSize),
			ctx:       ctx,
			cancel:    cancel,
		}
	})
	return instance
}

// OnInit implements the launcher.Component interface to initialize the worker pool.
func (w *workerComponent) OnInit() error {
	w.logger.Info("worker: inicializando pool de sombras",
		"pool_size", w.cfg.PoolSize,
		"buffer_size", w.cfg.BufferSize)
	return nil
}

// OnStart implements the launcher.Component interface to start the background workers.
func (w *workerComponent) OnStart() error {
	w.logger.Info("worker: activando workers en segundo plano")
	for i := 0; i < w.cfg.PoolSize; i++ {
		w.wg.Add(1)
		go func(id int) {
			defer w.wg.Done()
			w.runWorker(id)
		}(i)
	}
	return nil
}

// OnStop implements the launcher.Component interface to gracefully shut down the worker pool,
// processing remaining tasks before exiting.
func (w *workerComponent) OnStop() error {
	w.logger.Info("worker: apagando pool, procesando tareas pendientes")
	close(w.taskQueue)
	w.cancel()
	w.wg.Wait()
	w.logger.Info("worker: pool de sombras cerrado correctamente")
	return nil
}

// Dispatch adds a task to the worker queue. It returns false if the queue is full.
func (w *workerComponent) Dispatch(task Task) bool {
	select {
	case w.taskQueue <- task:
		return true
	default:
		w.logger.Error("worker: sobrecarga - cola llena, tarea ignorada", nil)
		return false
	}
}

// runWorker is the internal loop for each worker goroutine.
func (w *workerComponent) runWorker(id int) {
	for task := range w.taskQueue {
		if err := task(w.ctx); err != nil {
			w.logger.Error("worker: error ejecutando tarea", err, "worker_id", id)
		}
	}
}
