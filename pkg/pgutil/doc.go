/*
Package pgutil provides a standard client and utilities for interacting with PostgreSQL.

It simplifies the initialization and usage of pgxpool, provides a consistent interface
for database operations, and maps PostgreSQL errors to standard application errors (apperr).

Features:
- Singleton database client with lifecycle management.
- Unit of Work (UOW) pattern for transaction management across multiple services.
- Standardized error handling (Unique violations, Not found, etc.).
- Convenient transaction management with rollback on error or panic.
- Configurable connection pool using environment variables.

Example (Client):

	cfg := &pgutil.Config{Host: "localhost", Port: 5432}
	db := pgutil.GetPostgresClient(logger, cfg)

	if err := db.OnInit(); err != nil {
		logger.Fatal("pgutil: fallo al inicializar bd", err)
	}

	err := db.GetExecutor(ctx).Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")

Example (Unit of Work):

	uow := pgutil.GetUnitOfWork(logger, db)
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
