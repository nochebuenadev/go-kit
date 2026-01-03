package dbutil

import (
	"context"

	"github.com/nochebuenadev/go-kit/pkg/health"
	"github.com/nochebuenadev/go-kit/pkg/launcher"
)

type (
	// Result summarizes an executed SQL command.
	Result interface {
		// RowsAffected returns the number of rows affected by the command.
		RowsAffected() (int64, error)
	}

	// Row is a single result row from a query.
	Row interface {
		// Scan copies the columns from the matched row into the values pointed at by dest.
		Scan(dest ...any) error
	}

	// Rows is an iterator over query results.
	Rows interface {
		// Next prepares the next result row for reading with the Scan method.
		Next() bool
		// Scan copies the columns in the current row into the values pointed at by dest.
		Scan(dest ...any) error
		// Close closes the Rows, preventing further enumeration.
		Close() error
		// Err returns the error, if any, that was encountered during iteration.
		Err() error
	}

	// Executor defines common SQL execution methods.
	Executor interface {
		// Exec runs a command that doesn't return rows (e.g. INSERT, UPDATE, DELETE).
		Exec(ctx context.Context, sql string, args ...any) (Result, error)
		// Query executes a query that is expected to return multiple rows.
		Query(ctx context.Context, sql string, args ...any) (Rows, error)
		// QueryRow executes a query that is expected to return at most one row.
		QueryRow(ctx context.Context, sql string, args ...any) Row
	}

	// Transaction represents a database transaction.
	Transaction interface {
		Executor
		// Commit commits the transaction.
		Commit(ctx context.Context) error
		// Rollback aborts the transaction.
		Rollback(ctx context.Context) error
	}

	// Provider manages the database connection and transaction lifecycle.
	Provider interface {
		// GetExecutor returns a DBExecutor based on the context. If a transaction is present
		// in the context (via UnitOfWork), it returns the transaction; otherwise, it returns the pool.
		GetExecutor(ctx context.Context) Executor
		// Begin starts and returns a new database transaction.
		Begin(ctx context.Context) (Transaction, error)
		// Ping verifies the connection to the database.
		Ping(ctx context.Context) error
	}

	// Component adds lifecycle and health management to the Provider.
	Component interface {
		launcher.Component
		health.Checkable
		Provider
	}
)
