package task_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
	core_postgres_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/postges/pool"
)

// --- mock Pool / Row / Rows / CommandTag -------------------------------------

type mockRow struct {
	scanFn func(dest ...any) error
}

func (r mockRow) Scan(dest ...any) error { return r.scanFn(dest...) }

type mockRows struct {
	scanFns []func(dest ...any) error
	idx     int
	closed  bool
	err     error
}

func (r *mockRows) Close()     { r.closed = true }
func (r *mockRows) Err() error { return r.err }
func (r *mockRows) Next() bool {
	r.idx++
	return r.idx <= len(r.scanFns)
}
func (r *mockRows) Scan(dest ...any) error { return r.scanFns[r.idx-1](dest...) }

type mockCommandTag struct {
	rowsAffected int64
}

func (c mockCommandTag) RowsAffected() int64 { return c.rowsAffected }

type mockPool struct {
	queryFn    func(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error)
	queryRowFn func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row
	execFn     func(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error)
}

func (p *mockPool) Query(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
	return p.queryFn(ctx, sql, args...)
}
func (p *mockPool) QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
	return p.queryRowFn(ctx, sql, args...)
}
func (p *mockPool) Exec(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
	return p.execFn(ctx, sql, args...)
}
func (p *mockPool) Close()                     {}
func (p *mockPool) OpTimeout() time.Duration   { return time.Second }

func strPtr(s string) *string { return &s }

// --- CreateTask -----------------------------------------------------------

func TestTaskRepository_CreateTask_Success(t *testing.T) {
	now := time.Now()
	var capturedArgs []any

	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, args ...any) core_postgres_pool.Row {
			capturedArgs = args
			return mockRow{scanFn: func(dest ...any) error {
				*dest[0].(*int) = 1
				*dest[1].(*int) = 1
				*dest[2].(*string) = "Buy milk"
				*dest[3].(**string) = nil
				*dest[4].(*bool) = false
				*dest[5].(*time.Time) = now
				*dest[6].(**time.Time) = nil
				*dest[7].(*int) = 7
				return nil
			}}
		},
	}
	repo := NewTaskRepository(pool)

	task := domain.NewTaskUninitialized("Buy milk", nil, 7)
	result, err := repo.CreateTask(context.Background(), task)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 || result.Title != "Buy milk" || result.AuthorUserID != 7 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(capturedArgs) != 6 {
		t.Fatalf("expected 6 args passed to QueryRow, got %d", len(capturedArgs))
	}
}

func TestTaskRepository_CreateTask_ForeignKeyViolation(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, _ ...any) core_postgres_pool.Row {
			return mockRow{scanFn: func(dest ...any) error {
				return fmt.Errorf("fk violation: %w", core_postgres_pool.ErrViolatesForeingKey)
			}}
		},
	}
	repo := NewTaskRepository(pool)

	task := domain.NewTaskUninitialized("Buy milk", nil, 999)
	_, err := repo.CreateTask(context.Background(), task)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for foreign key violation, got: %v", err)
	}
}

func TestTaskRepository_CreateTask_UnknownError(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, _ ...any) core_postgres_pool.Row {
			return mockRow{scanFn: func(dest ...any) error {
				return errors.New("connection reset")
			}}
		},
	}
	repo := NewTaskRepository(pool)

	_, err := repo.CreateTask(context.Background(), domain.NewTaskUninitialized("Buy milk", nil, 1))

	if err == nil {
		t.Fatalf("expected error")
	}
	if errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("did not expect ErrNotFound for a generic error")
	}
}

// --- GetTask -----------------------------------------------------------

func TestTaskRepository_GetTask_Success(t *testing.T) {
	now := time.Now()
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, args ...any) core_postgres_pool.Row {
			if len(args) != 1 || args[0].(int) != 5 {
				t.Fatalf("expected id=5 as query arg, got %v", args)
			}
			return mockRow{scanFn: func(dest ...any) error {
				*dest[0].(*int) = 5
				*dest[1].(*int) = 2
				*dest[2].(*string) = "Buy milk"
				*dest[3].(**string) = strPtr("2%")
				*dest[4].(*bool) = false
				*dest[5].(*time.Time) = now
				*dest[6].(**time.Time) = nil
				*dest[7].(*int) = 1
				return nil
			}}
		},
	}
	repo := NewTaskRepository(pool)

	task, err := repo.GetTask(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 5 || task.Version != 2 {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestTaskRepository_GetTask_NotFound(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, _ ...any) core_postgres_pool.Row {
			return mockRow{scanFn: func(dest ...any) error {
				return core_postgres_pool.ErrNoRows
			}}
		},
	}
	repo := NewTaskRepository(pool)

	_, err := repo.GetTask(context.Background(), 999)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

// --- GetTasks (dynamic WHERE clause) -----------------------------------------

func TestTaskRepository_GetTasks_NoUserIDFilter(t *testing.T) {
	var capturedSQL string
	var capturedArgs []any

	pool := &mockPool{
		queryFn: func(_ context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
			capturedSQL = sql
			capturedArgs = args
			return &mockRows{scanFns: nil}, nil // no rows
		},
	}
	repo := NewTaskRepository(pool)

	limit, offset := 10, 0
	tasks, err := repo.GetTasks(context.Background(), nil, &limit, &offset)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
	if strings.Contains(capturedSQL, "WHERE") {
		t.Fatalf("expected no WHERE clause when userID is nil, got SQL: %s", capturedSQL)
	}
	if len(capturedArgs) != 2 {
		t.Fatalf("expected 2 args (limit, offset), got %d: %v", len(capturedArgs), capturedArgs)
	}
}

func TestTaskRepository_GetTasks_WithUserIDFilter(t *testing.T) {
	now := time.Now()
	var capturedSQL string
	var capturedArgs []any

	pool := &mockPool{
		queryFn: func(_ context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
			capturedSQL = sql
			capturedArgs = args
			return &mockRows{scanFns: []func(dest ...any) error{
				func(dest ...any) error {
					*dest[0].(*int) = 1
					*dest[1].(*int) = 1
					*dest[2].(*string) = "Task 1"
					*dest[3].(**string) = nil
					*dest[4].(*bool) = false
					*dest[5].(*time.Time) = now
					*dest[6].(**time.Time) = nil
					*dest[7].(*int) = 3
					return nil
				},
				func(dest ...any) error {
					*dest[0].(*int) = 2
					*dest[1].(*int) = 1
					*dest[2].(*string) = "Task 2"
					*dest[3].(**string) = nil
					*dest[4].(*bool) = false
					*dest[5].(*time.Time) = now
					*dest[6].(**time.Time) = nil
					*dest[7].(*int) = 3
					return nil
				},
			}}, nil
		},
	}
	repo := NewTaskRepository(pool)

	userID, limit, offset := 3, 10, 0
	tasks, err := repo.GetTasks(context.Background(), &userID, &limit, &offset)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if !strings.Contains(capturedSQL, "WHERE author_user_id = $3") {
		t.Fatalf("expected WHERE author_user_id = $3 clause, got SQL: %s", capturedSQL)
	}
	if len(capturedArgs) != 3 || capturedArgs[2].(*int) != &userID {
		t.Fatalf("expected 3rd arg to be the userID pointer, got %v", capturedArgs)
	}
}

// --- DeleteTask -----------------------------------------------------------

func TestTaskRepository_DeleteTask_Success(t *testing.T) {
	pool := &mockPool{
		execFn: func(_ context.Context, _ string, _ ...any) (core_postgres_pool.CommandTag, error) {
			return mockCommandTag{rowsAffected: 1}, nil
		},
	}
	repo := NewTaskRepository(pool)

	if err := repo.DeleteTask(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskRepository_DeleteTask_NotFound(t *testing.T) {
	pool := &mockPool{
		execFn: func(_ context.Context, _ string, _ ...any) (core_postgres_pool.CommandTag, error) {
			return mockCommandTag{rowsAffected: 0}, nil
		},
	}
	repo := NewTaskRepository(pool)

	err := repo.DeleteTask(context.Background(), 999)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when 0 rows affected, got: %v", err)
	}
}

// --- PatchTask (optimistic locking) -----------------------------------------

func TestTaskRepository_PatchTask_Success(t *testing.T) {
	now := time.Now()
	var capturedArgs []any

	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, args ...any) core_postgres_pool.Row {
			capturedArgs = args
			return mockRow{scanFn: func(dest ...any) error {
				*dest[0].(*int) = 1
				*dest[1].(*int) = 2 // version bumped
				*dest[2].(*string) = "Updated title"
				*dest[3].(**string) = nil
				*dest[4].(*bool) = false
				*dest[5].(*time.Time) = now
				*dest[6].(**time.Time) = nil
				*dest[7].(*int) = 1
				return nil
			}}
		},
	}
	repo := NewTaskRepository(pool)

	task := domain.Task{ID: 1, Version: 1, Title: "Updated title", CreatedAt: now}
	result, err := repo.PatchTask(context.Background(), 1, task)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != 2 {
		t.Fatalf("expected version bumped to 2, got %d", result.Version)
	}
	// last two args should be id and the *old* version used for the WHERE clause
	if capturedArgs[len(capturedArgs)-2].(int) != 1 {
		t.Fatalf("expected id=1 in args, got %v", capturedArgs)
	}
	if capturedArgs[len(capturedArgs)-1].(int) != 1 {
		t.Fatalf("expected version=1 (pre-patch) in args, got %v", capturedArgs)
	}
}

func TestTaskRepository_PatchTask_VersionConflict(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, _ ...any) core_postgres_pool.Row {
			return mockRow{scanFn: func(dest ...any) error {
				return core_postgres_pool.ErrNoRows // no row matched id+version
			}}
		},
	}
	repo := NewTaskRepository(pool)

	task := domain.Task{ID: 1, Version: 1, Title: "Updated title", CreatedAt: time.Now()}
	_, err := repo.PatchTask(context.Background(), 1, task)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict on optimistic lock failure, got: %v", err)
	}
}
