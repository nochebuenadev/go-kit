package dbutil

import (
	"context"
	"errors"
	"testing"

	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

// mockExecutor implements Executor for testing.
type mockExecutor struct{}

func (m *mockExecutor) Exec(ctx context.Context, sql string, args ...any) (Result, error) {
	return nil, nil
}
func (m *mockExecutor) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return nil, nil
}
func (m *mockExecutor) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return nil
}

// mockTransaction implements Transaction for testing.
type mockTransaction struct {
	mockExecutor
	CommitCalled   bool
	RollbackCalled bool
	CommitErr      error
}

func (m *mockTransaction) Commit(ctx context.Context) error {
	m.CommitCalled = true
	return m.CommitErr
}

func (m *mockTransaction) Rollback(ctx context.Context) error {
	m.RollbackCalled = true
	return nil
}

// mockProvider implements Provider for testing.
type mockProvider struct {
	tx    Transaction
	txErr error
}

func (m *mockProvider) GetExecutor(ctx context.Context) Executor {
	return nil
}

func (m *mockProvider) Begin(ctx context.Context) (Transaction, error) {
	return m.tx, m.txErr
}

func (m *mockProvider) Ping(ctx context.Context) error {
	return nil
}

// mockLogger is a simple mock for logz.Logger.
type mockLogger struct {
	logz.Logger
}

func (m *mockLogger) Debug(msg string, args ...any)               {}
func (m *mockLogger) Info(msg string, args ...any)                {}
func (m *mockLogger) Warn(msg string, args ...any)                {}
func (m *mockLogger) Error(msg string, err error, args ...any)    {}
func (m *mockLogger) LogError(msg string, err error, args ...any) {}
func (m *mockLogger) Fatal(msg string, err error, args ...any)    {}
func (m *mockLogger) With(args ...any) logz.Logger                { return m }
func (m *mockLogger) WithContext(ctx context.Context) logz.Logger { return m }

func TestUnitOfWork_Do(t *testing.T) {
	logger := &mockLogger{}

	t.Run("success", func(t *testing.T) {
		mtx := &mockTransaction{}
		mdb := &mockProvider{tx: mtx}
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
		mdb := &mockProvider{txErr: expectedErr}
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

	t.Run("function failure", func(t *testing.T) {
		mtx := &mockTransaction{}
		mdb := &mockProvider{tx: mtx}
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

	t.Run("panic handling", func(t *testing.T) {
		mtx := &mockTransaction{}
		mdb := &mockProvider{tx: mtx}
		uow := &unitOfWork{logger: logger, client: mdb}

		func() {
			defer func() {
				if r := recover(); r == nil {
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
