package pgutil

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
)

// mockTx implements a minimal pgx.Tx for testing.
type mockTx struct {
	pgx.Tx
	CommitCalled   bool
	RollbackCalled bool
	CommitErr      error
}

func (m *mockTx) Commit(ctx context.Context) error {
	m.CommitCalled = true
	return m.CommitErr
}

func (m *mockTx) Rollback(ctx context.Context) error {
	m.RollbackCalled = true
	return nil
}

// mockDBProvider implements DBProvider for testing.
type mockDBProvider struct {
	DBProvider
	tx    pgx.Tx
	txErr error
}

func (m *mockDBProvider) GetTRX(ctx context.Context) (pgx.Tx, error) {
	return m.tx, m.txErr
}

func TestUnitOfWork_Do(t *testing.T) {
	logger := &mockLogger{}

	t.Run("success", func(t *testing.T) {
		mtx := &mockTx{}
		mdb := &mockDBProvider{tx: mtx}
		uow := GetUnitOfWork(logger, mdb)

		err := uow.Do(context.Background(), func(ctx context.Context) error {
			tx, ok := TXFromContext(ctx)
			if !ok || tx != mtx {
				t.Error("expected transaction in context")
			}
			return nil
		})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !mtx.CommitCalled {
			t.Error("expected commit to be called")
		}
	})

	t.Run("start transaction failure", func(t *testing.T) {
		expectedErr := errors.New("start failed")
		mdb := &mockDBProvider{txErr: expectedErr}
		uow := &unitOfWork{logger: logger, client: mdb} // avoid singleton for fresh state

		err := uow.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		ae, ok := err.(*apperr.AppErr)
		if !ok || ae.GetCode() != "INTERNAL_ERROR" {
			t.Errorf("expected INTERNAL_ERROR, got %v", err)
		}
	})

	t.Run("function failure", func(t *testing.T) {
		mtx := &mockTx{}
		mdb := &mockDBProvider{tx: mtx}
		uow := &unitOfWork{logger: logger, client: mdb}
		expectedErr := errors.New("func failed")

		err := uow.Do(context.Background(), func(ctx context.Context) error {
			return expectedErr
		})

		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
		if !mtx.RollbackCalled {
			t.Error("expected rollback to be called")
		}
		if mtx.CommitCalled {
			t.Error("expected commit not to be called")
		}
	})

	t.Run("commit failure", func(t *testing.T) {
		commitErr := errors.New("commit failed")
		mtx := &mockTx{CommitErr: commitErr}
		mdb := &mockDBProvider{tx: mtx}
		uow := &unitOfWork{logger: logger, client: mdb}

		err := uow.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		ae, ok := err.(*apperr.AppErr)
		if !ok || ae.GetCode() != "INTERNAL_ERROR" {
			t.Errorf("expected INTERNAL_ERROR, got %v", err)
		}
	})

	t.Run("panic handling", func(t *testing.T) {
		mtx := &mockTx{}
		mdb := &mockDBProvider{tx: mtx}
		uow := &unitOfWork{logger: logger, client: mdb}

		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Errorf("expected panic to be propagated")
				}
			}()

			_ = uow.Do(context.Background(), func(ctx context.Context) error {
				panic("pánico de prueba")
			})
		}()

		if !mtx.RollbackCalled {
			t.Errorf("expected rollback to be called on panic")
		}
	})
}
