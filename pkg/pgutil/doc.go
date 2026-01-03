/*
Package pgutil provides a standard client and utilities for interacting with PostgreSQL.

It implements the interfaces defined in dbutil, simplifying the initialization
and usage of pgxpool, and maps PostgreSQL errors onto standard application errors (apperr).

Features:
- Singleton database client with dbutil.Component implementation.
- Specialized wrappers for pgx.Row, pgx.Rows, and pgx.Tx to satisfy agnostic interfaces.
- Standardized error handling (Unique violations, integrity violations, etc.).
- Configurable connection pool using environment variables.

Example (Client):

	cfg := &pgutil.Config{Host: "localhost", Port: 5432}
	db := pgutil.GetPostgresClient(logger, cfg) // returns dbutil.Component

	if err := db.OnInit(); err != nil {
		logger.Fatal("pgutil: fallo al inicializar bd", err)
	}

	err := db.GetExecutor(ctx).Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")

Example (Unit of Work):

	uow := dbutil.GetUnitOfWork(logger, db)
	err := uow.Do(ctx, func(ctx context.Context) error {
		// All operations inside this function use the same transaction
		executor := db.GetExecutor(ctx)
		if err := executor.Exec(ctx, "INSERT INTO table1 ..."); err != nil {
			return err
		}
		return executor.Exec(ctx, "INSERT INTO table2 ...")
	})
*/
package pgutil
