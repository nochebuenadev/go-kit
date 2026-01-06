// Package httputil provides a resilient and observable HTTP client.
//
// It encapsulates standard library http.Client with additional layers of
// resilience (Retries, Circuit Breaking) and observability (Logging, Tracing).
//
// Key Features:
// - Circuit Breaker: Automatically "opens" after a configurable threshold of failures.
// - Retries: Automatic retries with Exponential Backoff for 5xx and network errors.
// - Observability: Automatic logging of Method, URL, Status, and Latency.
// - Tracing: Automatic propagation of X-Request-ID.
// - Generic Helpers: Type-safe JSON decoding with DoJSON[T] helper.
// - Error Mapping: Automatic mapping of HTTP status codes to application errors (apperr).
//
// Example (Simple):
//
//	client := httputil.GetClient(logger, nil)
//	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/data", nil)
//	resp, err := client.Do(req)
//
// Example (Generic JSON):
//
//	data, err := httputil.DoJSON[MyType](ctx, client, req)
package httputil
