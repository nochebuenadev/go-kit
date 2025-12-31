/*
Package worker provides a background task execution pool (Shadow Workers).

It allows offloading long-running or non-blocking tasks to a pool of concurrent workers,
ensuring that the main application flow remains responsive. It includes built-in
backpressure handling and graceful shutdown.

Features:
- Configurable pool size and buffer capacity.
- Thread-safe task dispatching.
- Graceful shutdown: Processes remaining tasks in the queue before exit.
- Integration with launcher.Component for lifecycle management.

Example usage:

	cfg := &worker.Config{PoolSize: 10, BufferSize: 100}
	wk := worker.GetWorker(cfg, logger)

	// In a handler or service:
	wk.Dispatch(func(ctx context.Context) error {
		// Do background work here
		return nil
	})
*/
package worker
