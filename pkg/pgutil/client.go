package pgutil

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// Row matches pgx.Row interface.
	Row pgx.Row
	// Rows matches pgx.Rows interface.
	Rows pgx.Rows
	// Tx matches pgx.Tx interface.
	Tx pgx.Tx

	// DBExecutor defines the set of operations that can be performed against the database.
	DBExecutor interface {
		// Execute runs a command that doesn't return rows (e.g. INSERT, UPDATE, DELETE).
		Execute(ctx context.Context, query string, args ...any) error
		// QueryRow executes a query that is expected to return at most one row.
		QueryRow(ctx context.Context, query string, args ...any) Row
		// Query executes a query that is expected to return multiple rows.
		Query(ctx context.Context, query string, args ...any) (Rows, error)
		// WithTransaction executes the given function within a database transaction.
		WithTransaction(ctx context.Context, fn func(Tx) error) error
		// Ping verifies the connection to the database.
		Ping(ctx context.Context) error
	}

	// DBComponent extends DBExecutor with lifecycle management methods.
	DBComponent interface {
		DBExecutor
		// OnInit initializes the database component.
		OnInit() error
		// OnStart starts the database component services.
		OnStart() error
		// OnStop stops the database component services and closes connections.
		OnStop() error
	}

	// postgresClient is the concrete implementation of DBComponent using pgxpool.
	postgresClient struct {
		logger logz.Logger
		cfg    *DatabaseConfig
		pool   *pgxpool.Pool
		mu     sync.RWMutex
	}

	// errRow is a helper to return an error when a QueryRow fails early.
	errRow struct{ err error }
)

var (
	errPoolNotInitialized = errors.New("el pool de postgres no está inicializado")
	clientInstance        DBComponent
	clientOnce            sync.Once
	clientInitOnce        sync.Once
)

// GetPostgresClient returns the singleton instance of the database component.
func GetPostgresClient(config *DatabaseConfig, logger logz.Logger) DBComponent {
	clientOnce.Do(func() {
		clientInstance = &postgresClient{
			cfg:    config,
			logger: logger,
		}
	})
	return clientInstance
}

// OnInit implements DBComponent.
func (c *postgresClient) OnInit() error {
	var initErr error
	clientInitOnce.Do(func() {
		poolConfig, err := c.getPoolConfig()
		if err != nil {
			initErr = err
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			c.logger.Error("Error al inicializar el cliente de postgres", err)
			initErr = err
			return
		}
		c.pool = pool
	})
	return initErr
}

// OnStart implements DBComponent.
func (c *postgresClient) OnStart() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.pool == nil {
		return errPoolNotInitialized
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.pool.Ping(ctx)
}

// OnStop implements DBComponent.
func (c *postgresClient) OnStop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pool != nil {
		c.logger.Info("Cerrando el pool de conexiones de postgres")
		c.pool.Close()
		c.pool = nil
	}
	return nil
}

// Execute implements DBExecutor.
func (c *postgresClient) Execute(ctx context.Context, query string, args ...any) error {
	c.logger.Debug("Postgres: Execute", "query", query)
	pool, err := c.getInternalPool()
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, query, args...)
	return c.handleError(err)
}

// QueryRow implements DBExecutor.
func (c *postgresClient) QueryRow(ctx context.Context, query string, args ...any) Row {
	c.logger.Debug("Postgres: QueryRow", "query", query)
	pool, err := c.getInternalPool()
	if err != nil {
		return &errRow{err: err}
	}
	return pool.QueryRow(ctx, query, args...)
}

// Query implements DBExecutor.
func (c *postgresClient) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	c.logger.Debug("Postgres: Query", "query", query)
	pool, err := c.getInternalPool()
	if err != nil {
		return nil, err
	}
	rows, err := pool.Query(ctx, query, args...)
	return rows, c.handleError(err)
}

// Ping implements DBExecutor.
func (c *postgresClient) Ping(ctx context.Context) error {
	pool, err := c.getInternalPool()
	if err != nil {
		return err
	}
	return pool.Ping(ctx)
}

// WithTransaction implements DBExecutor.
func (c *postgresClient) WithTransaction(ctx context.Context, fn func(Tx) error) error {
	c.logger.Debug("Postgres: Iniciando transacción")
	pool, err := c.getInternalPool()
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return c.handleError(err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return c.handleError(fmt.Errorf("error en transacción: %v, el rollback falló: %v", err, rbErr))
		}
		return c.handleError(err)
	}

	return c.handleError(tx.Commit(ctx))
}

// Scan implements the Row interface for errRow, always returning the stored error.
func (e *errRow) Scan(dest ...any) error { return e.err }

// getInternalPool safely retrieves the underlying pgxpool.Pool instance.
// It returns an apperr.Internal if the pool is not initialized.
func (c *postgresClient) getInternalPool() (*pgxpool.Pool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.pool == nil {
		return nil, apperr.Internal("pool de base de datos no inicializado")
	}
	return c.pool, nil
}

// handleError maps PostgreSQL and pgx errors to standard application errors (apperr).
// It identifies unique violations, foreign key violations, and no-rows-found scenarios.
func (c *postgresClient) handleError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return apperr.New(apperr.ErrConflict, "el registro ya existe").
				WithContext("constraint", pgErr.ConstraintName).WithError(err)
		case pgerrcode.ForeignKeyViolation:
			return apperr.New(apperr.ErrInvalidInput, "violación de integridad de datos").
				WithContext("table", pgErr.TableName).WithError(err)
		}
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.NotFound("registro no encontrado").WithError(err)
	}

	return apperr.Internal("error inesperado en la base de datos").WithError(err)
}

// getPoolConfig parses the connection string and sets pool settings from configuration.
func (c *postgresClient) getPoolConfig() (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(c.cfg.GetConnectionString())
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = int32(c.cfg.MaxConns)
	poolConfig.MinConns = int32(c.cfg.MinConns)
	return poolConfig, nil
}
