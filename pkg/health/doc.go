/*
Package health provides a centralized health check mechanism for the application and its infrastructure.

It allows registering individual components (Checkable) that can report their status.
The health handler aggregates these results and provides an HTTP endpoint for observability.

Features:
- Level-based health reporting (Critical vs. Degraded).
- Concurrent execution: All checks run in parallel for faster responses.
- Automatic latency measurement for each check.
- Standardized JSON response format suitable for monitoring tools.
- Integration with external components like databases (pgutil) and caches (vkutil).

Example usage:

	// Create health handler with checks
	healthHandler := health.GetHandler(logger, dbClient, cacheClient)

	// Register in Echo
	e.GET("/health", healthHandler.HealthCheck)
*/
package health
