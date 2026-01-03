package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/logz"
	"github.com/sony/gobreaker"
)

type (
	// Client defines the interface for the resilient HTTP client.
	Client interface {
		// Do executes an HTTP request and returns the response.
		// It handles retries, circuit breaking, and logging automatically.
		Do(req *http.Request) (*http.Response, error)
	}

	// httpClient is the concrete implementation of the Client interface.
	httpClient struct {
		client *http.Client
		logger logz.Logger
		cfg    *Config
		cb     *gobreaker.CircuitBreaker
	}
)

var (
	// clientInstance is the singleton HTTP client.
	clientInstance Client
	// clientOnce ensures the client is initialized only once.
	clientOnce sync.Once
)

// GetClient returns the singleton instance of the resilient HTTP client.
func GetClient(logger logz.Logger, cfg *Config) Client {
	clientOnce.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
		}

		cbSettings := gobreaker.Settings{
			Name:        "http-client",
			MaxRequests: 0,
			Interval:    0,
			Timeout:     cfg.CBTimeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= cfg.CBThreshold
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				logger.Warn("httputil: cambio de estado del circuit breaker", "name", name, "from", from.String(), "to", to.String())
			},
		}

		clientInstance = &httpClient{
			client: &http.Client{
				Timeout: cfg.Timeout,
				Transport: &http.Transport{
					DialContext: (&net.Dialer{
						Timeout: cfg.DialTimeout,
					}).DialContext,
				},
			},
			logger: logger,
			cfg:    cfg,
			cb:     gobreaker.NewCircuitBreaker(cbSettings),
		}
	})

	return clientInstance
}

// Do executes the request with retries and circuit breaking.
func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response

	// Execute with Circuit Breaker
	result, err := c.cb.Execute(func() (any, error) {
		var innerErr error

		// Execute with Retries
		err := retry.Do(
			func() error {
				// Inject X-Request-ID for tracing
				if requestID := c.getRequestID(req.Context()); requestID != "" {
					req.Header.Set("X-Request-ID", requestID)
				}

				start := time.Now()
				resp, innerErr = c.client.Do(req)
				latency := time.Since(start)

				if innerErr != nil {
					c.logger.Debug("httputil: error en petición de red", "err", innerErr, "url", req.URL.String())
					return innerErr
				}

				c.logMetadata(req, resp, latency)

				// Only retry on 5xx errors
				if resp.StatusCode >= 500 {
					return fmt.Errorf("server error: %d", resp.StatusCode)
				}

				return nil
			},
			retry.Attempts(c.cfg.MaxRetries),
			retry.Delay(c.cfg.RetryDelay),
			retry.DelayType(retry.BackOffDelay), // Fixed: use DelayType for backoff strategy
			retry.LastErrorOnly(true),
			retry.RetryIf(func(err error) bool {
				// Retry on network errors or 5xx
				return err != nil
			}),
		)

		return resp, err
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return nil, apperr.Internal("servicio externo no disponible (circuito abierto)").WithError(err)
		}
		return nil, c.mapError(err, resp)
	}

	return result.(*http.Response), nil
}

// DoJSON is a generic helper to execute a request and decode the JSON response.
func DoJSON[T any](ctx context.Context, client Client, req *http.Request) (*T, error) {
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, MapStatusToError(resp.StatusCode, "error en respuesta de API externa")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperr.Internal("error al leer el cuerpo de la respuesta").WithError(err)
	}

	var data T
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, apperr.Internal("error al decodificar respuesta JSON").WithError(err)
	}

	return &data, nil
}

// MapStatusToError converts an HTTP status code to its corresponding apperr.
func MapStatusToError(code int, msg string) error {
	switch code {
	case http.StatusNotFound:
		return apperr.NotFound("%s", msg)
	case http.StatusBadRequest:
		return apperr.InvalidInput("%s", msg)
	case http.StatusUnauthorized:
		return apperr.New(apperr.ErrUnauthorized, msg)
	case http.StatusForbidden:
		return apperr.New(apperr.ErrPermissionDenied, msg)
	case http.StatusConflict:
		return apperr.New(apperr.ErrConflict, msg)
	case http.StatusTooManyRequests:
		return apperr.New(apperr.ErrUnavailable, msg)
	default:
		return apperr.Internal("%s", msg)
	}
}

// logMetadata logs request and response metadata.
func (c *httpClient) logMetadata(req *http.Request, resp *http.Response, latency time.Duration) {
	c.logger.Info("httputil: petición completada",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"latency", latency.String(),
	)

	// Dump headers and body if debug mode is active
	dump, err := httputil.DumpResponse(resp, true)
	if err == nil {
		c.logger.Debug("httputil: volcado de respuesta", "dump", string(dump))
	}
}

// getRequestID retrieves the X-Request-ID from the context or logs if available.
func (c *httpClient) getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}

// mapError converts various errors to apperr.
func (c *httpClient) mapError(err error, resp *http.Response) error {
	if resp != nil {
		return MapStatusToError(resp.StatusCode, "error en petición HTTP")
	}
	return apperr.Internal("error de red o timeout").WithError(err)
}
