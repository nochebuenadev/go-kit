/*
Package pgutil provides a standard client and utilities for interacting with PostgreSQL.

It simplifies the initialization and usage of pgxpool, provides a consistent interface
for database operations, and maps PostgreSQL errors to standard application errors (apperr).

Features:
- Singleton database client with lifecycle management.
- Standardized error handling (Unique violations, Not found, etc.).
- Convenient transaction management with rollback on error or panic.
- Configurable connection pool using environment variables.

Example:

	cfg := &pgutil.DatabaseConfig{Host: "localhost", Port: 5432, ...}
	db := pgutil.GetPostgresClient(cfg, logger)

	if err := db.OnInit(); err != nil {
		logger.Fatal("failed to init db", err)
	}

	err := db.Execute(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
*/
package pgutil
