package pgutil

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/dbutil"
	"github.com/nochebuenadev/go-kit/pkg/health"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// pgComponent is the concrete implementation of dbutil.Component using pgxpool.
	pgComponent struct {
		// logger is used for tracking database operations and errors.
		logger logz.Logger
		// cfg is the database connection configuration.
		cfg *Config
		// pool is the underlying pgx connection pool.
		pool *pgxpool.Pool
		// mu protects access to the pool instance during lifecycle changes.
		mu sync.RWMutex
	}

	// pgResult wraps pgconn.CommandTag to satisfy the dbutil.Result interface.
	pgResult struct {
		// tag is the underlying pgx result tag.
		tag pgconn.CommandTag
	}

	// pgRows wraps pgx.Rows to satisfy the dbutil.Rows interface.
	pgRows struct {
		pgx.Rows
	}

	// pgRow wraps pgx.Row to satisfy the dbutil.Row interface.
	pgRow struct {
		pgx.Row
	}

	// pgTx wraps pgx.Tx to satisfy the dbutil.Transaction interface.
	pgTx struct {
		pgx.Tx
	}

	// errRow is an implementation of dbutil.Row that always returns a stored error.
	// Used for handling early query failures gracefully.
	errRow struct {
		// err is the stored error to return on Scan.
		err error
	}
)

var (
	// errPoolNotInitialized is returned when an operation is attempted before OnInit.
	errPoolNotInitialized = errors.New("el pool de postgres no está inicializado")
	// clientInstance is the singleton database component.
	clientInstance dbutil.Component
	// clientOnce ensures the component is initialized only once.
	clientOnce sync.Once
	// clientInitOnce ensures the connection pool is created only once.
	clientInitOnce sync.Once
)

// GetPostgresClient returns the singleton instance of the database component.
func GetPostgresClient(logger logz.Logger, config *Config) dbutil.Component {
	clientOnce.Do(func() {
		clientInstance = &pgComponent{
			cfg:    config,
			logger: logger,
		}
	})
	return clientInstance
}

// HandleError maps PostgreSQL and pgx errors to standard application errors (apperr).
// It identifies unique violations, foreign key violations, and no-rows-found scenarios.
func HandleError(err error) error {
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

// OnInit implements the launcher.Component interface to initialize the database pool.
func (c *pgComponent) OnInit() error {
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
			c.logger.Error("pgutil: error al inicializar el cliente de Postgres", err)
			initErr = err
			return
		}
		c.pool = pool
	})
	return initErr
}

// OnStart implements the launcher.Component interface to verify the database connection.
func (c *pgComponent) OnStart() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.pool == nil {
		return errPoolNotInitialized
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.pool.Ping(ctx)
}

// OnStop implements the launcher.Component interface to gracefully close the database pool.
func (c *pgComponent) OnStop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pool != nil {
		c.logger.Info("pgutil: cerrando el pool de conexiones de Postgres")
		c.pool.Close()
		c.pool = nil
	}
	return nil
}

// Ping implements dbutil.Provider.
func (c *pgComponent) Ping(ctx context.Context) error {
	pool, err := c.getInternalPool()
	if err != nil {
		return err
	}
	return pool.Ping(ctx)
}

func (c *pgComponent) Exec(ctx context.Context, sql string, args ...any) (dbutil.Result, error) {
	tag, err := c.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgResult{tag: tag}, nil
}

func (c *pgComponent) Query(ctx context.Context, sql string, args ...any) (dbutil.Rows, error) {
	rows, err := c.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgRows{Rows: rows}, nil
}

// QueryRow implements dbutil.Executor.
func (c *pgComponent) QueryRow(ctx context.Context, sql string, args ...any) dbutil.Row {
	return &pgRow{Row: c.pool.QueryRow(ctx, sql, args...)}
}

// GetExecutor implements dbutil.Provider.
// It retrieves the active transaction from the context if it exists,
// allowing for transparent execution within the Unit of Work.
func (c *pgComponent) GetExecutor(ctx context.Context) dbutil.Executor {
	if tx, ok := dbutil.TXFromContext(ctx); ok {
		return tx
	}
	return c
}

// Begin implements dbutil.Provider.
func (c *pgComponent) Begin(ctx context.Context) (dbutil.Transaction, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgTx{Tx: tx}, nil
}

// RowsAffected implements dbutil.Result.
func (r *pgResult) RowsAffected() (int64, error) {
	return r.tag.RowsAffected(), nil
}

// Exec implements dbutil.Executor.
func (t *pgTx) Exec(ctx context.Context, sql string, args ...any) (dbutil.Result, error) {
	tag, err := t.Tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgResult{tag: tag}, nil
}

// Query implements dbutil.Executor.
func (t *pgTx) Query(ctx context.Context, sql string, args ...any) (dbutil.Rows, error) {
	rows, err := t.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgRows{Rows: rows}, nil
}

// QueryRow implements dbutil.Executor.
func (t *pgTx) QueryRow(ctx context.Context, sql string, args ...any) dbutil.Row {
	return &pgRow{Row: t.Tx.QueryRow(ctx, sql, args...)}
}

// Commit implements dbutil.Transaction.
func (t *pgTx) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

// Rollback implements dbutil.Transaction.
func (t *pgTx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}

// Close implements dbutil.Rows.
func (r *pgRows) Close() error {
	r.Rows.Close()
	return nil
}

// Scan implements the Row interface for errRow, always returning the stored error.
func (e *errRow) Scan(dest ...any) error { return e.err }

// getInternalPool safely retrieves the underlying pgxpool.Pool instance.
// It returns an apperr.Internal if the pool is not initialized.
func (c *pgComponent) getInternalPool() (*pgxpool.Pool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.pool == nil {
		return nil, apperr.Internal("pool de base de datos no inicializado")
	}
	return c.pool, nil
}

// HealthCheck implements the health.Checkable interface.
func (c *pgComponent) HealthCheck(ctx context.Context) error { return c.Ping(ctx) }

// Name implements the health.Checkable interface.
func (c *pgComponent) Name() string { return "postgres" }

// Priority implements the health.Checkable interface.
func (c *pgComponent) Priority() health.Level { return health.LevelCritical }

// getPoolConfig parses the connection string and sets pool settings from configuration.
func (c *pgComponent) getPoolConfig() (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(c.cfg.GetConnectionString())
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = int32(c.cfg.MaxConns)
	poolConfig.MinConns = int32(c.cfg.MinConns)
	return poolConfig, nil
}
