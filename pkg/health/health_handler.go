package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// Level represents the criticality of a component to the overall application health.
	Level int

	// Checkable defines the interface that components must implement to be verified by the health handler.
	Checkable interface {
		// HealthCheck performs a connectivity or status check.
		HealthCheck(ctx context.Context) error
		// Name returns the human-readable name of the component.
		Name() string
		// Priority returns the criticality level of the component.
		Priority() Level
	}

	// ComponentStatus represents the health state of an individual component.
	ComponentStatus struct {
		// Status is the health state (UP, DEGRADED, DOWN).
		Status string `json:"status"`
		// Latency is the time taken to perform the check.
		Latency string `json:"latency,omitempty"`
		// Error is the error message if the check failed.
		Error string `json:"error,omitempty"`
	}

	// Response represents the overall health check response body.
	Response struct {
		// Status is the overall health state of the application.
		Status string `json:"status"`
		// Components is a map of individual component statuses.
		Components map[string]ComponentStatus `json:"components"`
	}

	// Handler defines the interface for the health check HTTP handler.
	Handler interface {
		// HealthCheck is the Echo handler function for the health check endpoint.
		HealthCheck(c echo.Context) error
	}

	// handler is the concrete implementation of the health Handler.
	handler struct {
		// logger is used for reporting health check requests.
		logger logz.Logger
		// checks is the list of components to verify.
		checks []Checkable
	}
)

const (
	// LevelCritical indicates that the component is essential for the application to function.
	// If a critical component is DOWN, the overall status will be DOWN.
	LevelCritical Level = iota
	// LevelDegraded indicates that the component is important but not essential.
	// If a degraded component is DOWN, the overall status will be DEGRADED.
	LevelDegraded
)

var (
	// handlerInstance is the singleton health handler.
	handlerInstance Handler
	// handlerOnce ensures the handler is initialized only once.
	handlerOnce sync.Once
)

// GetHandler returns the singleton instance of the health handler.
// checks is a list of components to verify during the health check.
func GetHandler(logger logz.Logger, checks ...Checkable) Handler {
	handlerOnce.Do(func() {
		handlerInstance = &handler{
			logger: logger,
			checks: checks,
		}
	})

	return handlerInstance
}

// HealthCheck executes all registered health checks concurrently and returns a summary response.
// It leverages goroutines to perform checks in parallel, respecting a global timeout.
// It returns HTTP 200 (OK) if all critical components are UP or DEGRADED,
// and HTTP 503 (Service Unavailable) if any critical component is DOWN.
func (h *handler) HealthCheck(c echo.Context) error {
	// El timeout global sigue siendo bueno como "seguro de vida"
	logger := h.logger.WithContext(c.Request().Context())
	logger.Debug("Health check request")

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	overallStatus := "UP"
	httpStatus := http.StatusOK

	type result struct {
		name     string
		status   ComponentStatus
		priority Level
	}

	resChan := make(chan result, len(h.checks))

	for _, check := range h.checks {
		go func(chk Checkable) {
			start := time.Now()
			err := chk.HealthCheck(ctx)
			latency := time.Since(start).String()

			status := "UP"
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
				status = "DOWN"
				if chk.Priority() == LevelDegraded {
					status = "DEGRADED"
				}
			}

			resChan <- result{
				name:     chk.Name(),
				priority: chk.Priority(),
				status: ComponentStatus{
					Status:  status,
					Latency: latency,
					Error:   errMsg,
				},
			}
		}(check)
	}

	components := make(map[string]ComponentStatus)

	for i := 0; i < len(h.checks); i++ {
		res := <-resChan
		components[res.name] = res.status

		if res.status.Status == "DOWN" {
			overallStatus = "DOWN"
			httpStatus = http.StatusServiceUnavailable
		} else if res.status.Status == "DEGRADED" && overallStatus == "UP" {
			overallStatus = "DEGRADED"
		}
	}

	return c.JSON(httpStatus, Response{
		Status:     overallStatus,
		Components: components,
	})
}
