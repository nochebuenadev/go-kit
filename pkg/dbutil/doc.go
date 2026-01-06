// Package dbutil provides a database-agnostic utility layer for SQL execution.
// It defines core interfaces that allow services to remain independent of the
// underlying database engine (e.g., PostgreSQL, MySQL).
//
// The package also implements the Unit of Work pattern to manage transactions
// across multiple operations in a consistent way.
package dbutil
