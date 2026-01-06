// Package mysqlutil provides a standard client and utilities for interacting with MySQL.
//
// It implements the interfaces defined in dbutil, simplifying the initialization
// and usage of database/sql with a MySQL driver, and maps MySQL errors onto
// standard application errors (apperr).
//
// Features:
// - Singleton database client with dbutil.Component implementation.
// - Specialized wrappers for sql.Row, sql.Rows, and sql.Tx to satisfy agnostic interfaces.
// - Standardized error handling (Duplicate entries, integrity violations, etc.).
// - Configurable connection pool using environment variables.
//
// Example:
//
//	cfg := &mysqlutil.Config{Host: "localhost", Port: 3306}
//	db := mysqlutil.GetMySQLClient(logger, cfg)
//
//	if err := db.OnInit(); err != nil {
//		logger.Fatal("mysqlutil: fallo al inicializar bd", err)
//	}
//
//	executor := db.GetExecutor(ctx)
//	var name string
//	err := executor.QueryRow(ctx, "SELECT name FROM users WHERE id = ?", 1).Scan(&name)
package mysqlutil
