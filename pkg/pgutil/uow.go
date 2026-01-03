package pgutil

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
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
		client DBProvider
	}
)

var (
	// uowInstance is the singleton unit of work instance.
	uowInstance UnitOfWork
	// uowOnce ensures that the unit of work is initialized only once.
	uowOnce sync.Once
)

// GetUnitOfWork returns the singleton instance of the Unit of Work.
func GetUnitOfWork(logger logz.Logger, client DBProvider) UnitOfWork {
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
func TXFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(uowKey{}).(pgx.Tx)
	return tx, ok
}

// Do implements the UnitOfWork interface.
// It handles the transaction lifecycle: begin, commit, and rollback.
func (uow *unitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	uow.logger.Debug("pgutil: iniciando transacción de unidad de trabajo")
	tx, err := uow.client.GetTRX(ctx)
	if err != nil {
		return apperr.Internal("error al iniciar transacción").WithError(err)
	}

	defer func() {
		if r := recover(); r != nil {
			uow.logger.Warn("pgutil: pánico detectado, revirtiendo transacción", "panic", r)
			_ = tx.Rollback(ctx)
			panic(r)
		}
		_ = tx.Rollback(ctx)
	}()

	ctxWithTX := setTXInContext(ctx, tx)

	if err := fn(ctxWithTX); err != nil {
		uow.logger.Debug("pgutil: función falló, revirtiendo transacción", "error", err)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		uow.logger.Error("pgutil: error al confirmar transacción", err)
		return apperr.Internal("error al confirmar transacción").WithError(err)
	}

	uow.logger.Debug("pgutil: transacción de unidad de trabajo confirmada correctamente")
	return nil
}

// setTXInContext injects a transaction into the context.
func setTXInContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, uowKey{}, tx)
}
