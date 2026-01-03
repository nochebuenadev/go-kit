package mysqlutil

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/dbutil"
	"github.com/nochebuenadev/go-kit/pkg/health"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// mysqlComponent is the concrete implementation of dbutil.Component using database/sql.
	mysqlComponent struct {
		logger logz.Logger
		cfg    *Config
		db     *sql.DB
		mu     sync.RWMutex
	}

	// mysqlResult wraps sql.Result to satisfy the dbutil.Result interface.
	mysqlResult struct {
		res sql.Result
	}

	// mysqlRows wraps sql.Rows to satisfy the dbutil.Rows interface.
	mysqlRows struct {
		*sql.Rows
	}

	// mysqlRow wraps sql.Row to satisfy the dbutil.Row interface.
	mysqlRow struct {
		*sql.Row
	}

	// mysqlTx wraps sql.Tx to satisfy the dbutil.Transaction interface.
	mysqlTx struct {
		*sql.Tx
	}
)

var (
	// errDBNotInitialized is returned when an operation is attempted before OnInit.
	errDBNotInitialized = errors.New("el cliente de MySQL no está inicializado")
	// clientInstance is the singleton database component.
	clientInstance dbutil.Component
	// clientOnce ensures the component is initialized only once.
	clientOnce sync.Once
	// clientInitOnce ensures the connection pool is created only once.
	clientInitOnce sync.Once
)

// GetMySQLClient returns the singleton instance of the MySQL database component.
func GetMySQLClient(logger logz.Logger, config *Config) dbutil.Component {
	clientOnce.Do(func() {
		clientInstance = &mysqlComponent{
			cfg:    config,
			logger: logger,
		}
	})
	return clientInstance
}

// HandleError maps MySQL errors to standard application errors (apperr).
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062: // ER_DUP_ENTRY
			return apperr.New(apperr.ErrConflict, "el registro ya existe").WithError(err)
		case 1216, 1217, 1451, 1452: // Foreign key violations
			return apperr.New(apperr.ErrInvalidInput, "violación de integridad de datos").WithError(err)
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return apperr.NotFound("registro no encontrado").WithError(err)
	}

	return apperr.Internal("error inesperado en la base de datos MySQL").WithError(err)
}

// OnInit implements the launcher.Component interface to initialize the database pool.
func (c *mysqlComponent) OnInit() error {
	var initErr error
	clientInitOnce.Do(func() {
		db, err := sql.Open("mysql", c.cfg.GetConnectionString())
		if err != nil {
			c.logger.Error("mysqlutil: error al abrir la conexión de MySQL", err)
			initErr = err
			return
		}

		db.SetMaxOpenConns(c.cfg.MaxConns)
		db.SetMaxIdleConns(c.cfg.MinConns)

		if duration, err := time.ParseDuration(c.cfg.MaxConnLifetime); err == nil {
			db.SetConnMaxLifetime(duration)
		}
		if duration, err := time.ParseDuration(c.cfg.MaxConnIdleTime); err == nil {
			db.SetConnMaxIdleTime(duration)
		}

		c.db = db
	})
	return initErr
}

// OnStart implements the launcher.Component interface to verify the database connection.
func (c *mysqlComponent) OnStart() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return errDBNotInitialized
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.db.PingContext(ctx)
}

// OnStop implements the launcher.Component interface to gracefully close the database pool.
func (c *mysqlComponent) OnStop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db != nil {
		c.logger.Info("mysqlutil: cerrando el pool de conexiones de MySQL")
		_ = c.db.Close()
		c.db = nil
	}
	return nil
}

// Ping implements dbutil.Provider.
func (c *mysqlComponent) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.db == nil {
		return errDBNotInitialized
	}
	return c.db.PingContext(ctx)
}

// Exec implements dbutil.Executor.
func (c *mysqlComponent) Exec(ctx context.Context, query string, args ...any) (dbutil.Result, error) {
	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &mysqlResult{res: res}, nil
}

// Query implements dbutil.Executor.
func (c *mysqlComponent) Query(ctx context.Context, query string, args ...any) (dbutil.Rows, error) {
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &mysqlRows{Rows: rows}, nil
}

// QueryRow implements dbutil.Executor.
func (c *mysqlComponent) QueryRow(ctx context.Context, query string, args ...any) dbutil.Row {
	return &mysqlRow{Row: c.db.QueryRowContext(ctx, query, args...)}
}

// GetExecutor implements dbutil.Provider.
func (c *mysqlComponent) GetExecutor(ctx context.Context) dbutil.Executor {
	if tx, ok := dbutil.TXFromContext(ctx); ok {
		return tx
	}
	return c
}

// Begin implements dbutil.Provider.
func (c *mysqlComponent) Begin(ctx context.Context) (dbutil.Transaction, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &mysqlTx{Tx: tx}, nil
}

// RowsAffected implements dbutil.Result.
func (r *mysqlResult) RowsAffected() (int64, error) {
	return r.res.RowsAffected()
}

// Exec implements dbutil.Executor.
func (t *mysqlTx) Exec(ctx context.Context, query string, args ...any) (dbutil.Result, error) {
	res, err := t.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &mysqlResult{res: res}, nil
}

// Query implements dbutil.Executor.
func (t *mysqlTx) Query(ctx context.Context, query string, args ...any) (dbutil.Rows, error) {
	rows, err := t.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &mysqlRows{Rows: rows}, nil
}

// QueryRow implements dbutil.Executor.
func (t *mysqlTx) QueryRow(ctx context.Context, query string, args ...any) dbutil.Row {
	return &mysqlRow{Row: t.Tx.QueryRowContext(ctx, query, args...)}
}

// Commit implements dbutil.Transaction.
func (t *mysqlTx) Commit(ctx context.Context) error {
	return t.Tx.Commit()
}

// Rollback implements dbutil.Transaction.
func (t *mysqlTx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback()
}

// Scan implements dbutil.Row.
func (r *mysqlRow) Scan(dest ...any) error {
	return r.Row.Scan(dest...)
}

// Next implements dbutil.Rows.
func (r *mysqlRows) Next() bool {
	return r.Rows.Next()
}

// Close implements dbutil.Rows.
func (r *mysqlRows) Close() error {
	return r.Rows.Close()
}

// Err implements dbutil.Rows.
func (r *mysqlRows) Err() error {
	return r.Rows.Err()
}

// HealthCheck implements health.Checkable.
func (c *mysqlComponent) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx)
}

// Name implements health.Checkable.
func (c *mysqlComponent) Name() string {
	return "mysql"
}

// Priority implements health.Checkable.
func (c *mysqlComponent) Priority() health.Level {
	return health.LevelCritical
}
