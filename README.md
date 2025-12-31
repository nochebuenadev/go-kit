# Go Kit

`go-kit` is a lightweight, opinionated production-ready toolkit for building scalable microservices in Go. It provides a standardized architecture based on managed lifecycles, structured observability, and consistent error handling.

## Core Features

### 🚀 Lifecycle Management ([pkg/launcher](file:///home/renenochebuena/Workspace/go-kit/pkg/launcher))
Managed application startup and shutdown. Components register `OnInit`, `OnStart`, and `OnStop` hooks to handle dependencies and graceful transitions.

### 🩺 Health Checks ([pkg/health](file:///home/renenochebuena/Workspace/go-kit/pkg/health))
Concurrent health monitoring for infrastructure dependencies. Supports `Critical` and `Degraded` levels with automatic latency measurement.

### 👥 Shadow Workers ([pkg/worker](file:///home/renenochebuena/Workspace/go-kit/pkg/worker))
A background task execution pool with configurable concurrency and backpressure handling.

### 📝 Structured Logging ([pkg/logz](file:///home/renenochebuena/Workspace/go-kit/pkg/logz))
Context-aware structured logging with correlation IDs for request tracing.

### ⚠️ Error Handling ([pkg/apperr](file:///home/renenochebuena/Workspace/go-kit/pkg/apperr))
Custom application errors with machine-readable codes, metadata, and automatic mapping to HTTP status codes.

### 🌐 Web & Middleware ([pkg/server](file:///home/renenochebuena/Workspace/go-kit/pkg/server), [pkg/mw](file:///home/renenochebuena/Workspace/go-kit/pkg/mw))
Echo-based HTTP server with pre-configured middlewares for logging, error recovery, and request tracing.

### 💾 Infrastructure ([pkg/pgutil](file:///home/renenochebuena/Workspace/go-kit/pkg/pgutil), [pkg/vkutil](file:///home/renenochebuena/Workspace/go-kit/pkg/vkutil))
Thread-safe, singleton-based clients for PostgreSQL (using `pgx`) and Valkey (using `valkey-go`).

---

## Getting Started

### Installation
```bash
go get github.com/nochebuenadev/go-kit
```

### Basic Example
```go
package main

import (
    "github.com/nochebuenadev/go-kit/pkg/launcher"
    "github.com/nochebuenadev/go-kit/pkg/logz"
    "github.com/nochebuenadev/go-kit/pkg/server"
)

func main() {
    logger := logz.MustInit("my-service", true)
    
    // Core Launcher
    l := launcher.New(logger)
    
    // HTTP Server Component
    srv := server.GetEchoServer(&server.Config{Port: 8080}, logger)
    l.Append(srv)

    // Run happily
    l.Run()
}
```

## Project Principles
1. **Singleton-First**: Infrastructure clients are managed as singletons to ensure safe resource reuse.
2. **Lifecycle Aware**: Every major component should implement the `launcher.Component` interface.
3. **Observability**: Logs and errors must carry correlation IDs and machine-readable context.
4. **Concurrency Safety**: Components like `health` and `worker` are designed for high-concurrency environments.

---

## Component Overview

| Package    | Description                                         |
|:-----------|:----------------------------------------------------|
| `apperr`   | Standardized error types and JSON marshaling.       |
| `check`    | Struct and field validation helpers.                |
| `health`   | Parallel health check aggregation and HTTP handler. |
| `launcher` | App lifecycle registry and signal handling.         |
| `logz`     | Slog-based structured logger with context support.  |
| `mw`       | Echo middlewares (Error Handler, Correlation ID).   |
| `pgutil`   | PostgreSQL connection pool and error mapping.       |
| `server`   | HTTP server setup and lifecycle management.         |
| `vkutil`   | Valkey (Redis compatible) client and health check.  |
| `worker`   | Background task pool with backpressure.             |

## Testing
The toolkit maintains high test coverage across all packages.
```bash
go test -v ./pkg/...
```
