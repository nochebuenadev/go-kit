package dbutil

import (
	"context"
	"sync"

	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// UnitOfWork defines the interface for the Unit of Work pattern,
	// ensuring that a set of operations are executed within a single transaction.
	UnitOfWork interface {
		// Do executes the provided function within a transaction.
		// If the function returns an error, the transaction is rolled back.
		// Otherwise, the transaction is committed.
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}

	// unitOfWork is the concrete implementation of UnitOfWork.
	unitOfWork struct {
		logger logz.Logger
		client Provider
	}

	// uowKey is a private type for storing the transaction in the context.
	uowKey struct{}
)

var (
	// uowInstance is the singleton unit of work instance.
	uowInstance UnitOfWork
	// uowOnce ensures that the unit of work is initialized only once.
	uowOnce sync.Once
)

// GetUnitOfWork returns the singleton instance of the Unit of Work.
func GetUnitOfWork(logger logz.Logger, client Provider) UnitOfWork {
	uowOnce.Do(func() {
		uowInstance = &unitOfWork{
			logger: logger,
			client: client,
		}
	})

	return uowInstance
}

// TXFromContext retrieves the active transaction from the context.
// It returns the transaction and true if found, nil and false otherwise.
func TXFromContext(ctx context.Context) (Transaction, bool) {
	tx, ok := ctx.Value(uowKey{}).(Transaction)
	return tx, ok
}

// Do implements the UnitOfWork interface.
// It handles the transaction lifecycle: begin, commit, and rollback.
func (uow *unitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	uow.logger.Debug("dbutil: iniciando transacción de unidad de trabajo")
	tx, err := uow.client.Begin(ctx)
	if err != nil {
		return apperr.Internal("error al iniciar transacción").WithError(err)
	}

	defer func() {
		if r := recover(); r != nil {
			uow.logger.Warn("dbutil: pánico detectado, revirtiendo transacción", "panic", r)
			_ = tx.Rollback(ctx)
			panic(r)
		}
		// Intentamos rollback por seguridad; si ya hubo commit, el driver lo ignorará.
		_ = tx.Rollback(ctx)
	}()

	ctxWithTX := setTXInContext(ctx, tx)

	if err := fn(ctxWithTX); err != nil {
		uow.logger.Debug("dbutil: función falló, revirtiendo transacción", "error", err)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		uow.logger.Error("dbutil: error al confirmar transacción", err)
		return apperr.Internal("error al confirmar transacción").WithError(err)
	}

	uow.logger.Debug("dbutil: transacción de unidad de trabajo confirmada correctamente")
	return nil
}

// setTXInContext injects a transaction into the context.
func setTXInContext(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, uowKey{}, tx)
}
