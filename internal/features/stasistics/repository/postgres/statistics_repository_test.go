package postgres_statistics_repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	core_postgres_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/postges/pool"
)

// --- mock Pool / Rows -------------------------------------------------------

type mockRows struct {
	scanFns []func(dest ...any) error
	idx     int
}

func (r *mockRows) Close()     {}
func (r *mockRows) Err() error { return nil }
func (r *mockRows) Next() bool {
	r.idx++
	return r.idx <= len(r.scanFns)
}
func (r *mockRows) Scan(dest ...any) error { return r.scanFns[r.idx-1](dest...) }

type mockPool struct {
	queryFn func(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error)
}

func (p *mockPool) Query(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
	return p.queryFn(ctx, sql, args...)
}
func (p *mockPool) QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
	panic("not used by GetTasks")
}
func (p *mockPool) Exec(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
	panic("not used by GetTasks")
}
func (p *mockPool) Close()                   {}
func (p *mockPool) OpTimeout() time.Duration { return time.Second }

func taskScanFn(id int) func(dest ...any) error {
	now := time.Now()
	return func(dest ...any) error {
		*dest[0].(*int) = id
		*dest[1].(*int) = 1
		*dest[2].(*string) = "Task"
		*dest[3].(**string) = nil
		*dest[4].(*bool) = false
		*dest[5].(*time.Time) = now
		*dest[6].(**time.Time) = nil
		*dest[7].(*int) = 1
		return nil
	}
}

// --- GetTasks (dynamic WHERE clause with up to 3 conditions) -----------------

func TestStatisticsRepository_GetTasks_NoFilters(t *testing.T) {
	var capturedSQL string
	var capturedArgs []any

	pool := &mockPool{
		queryFn: func(_ context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
			capturedSQL, capturedArgs = sql, args
			return &mockRows{}, nil
		},
	}
	repo := NewStatisticsRepository(pool)

	tasks, err := repo.GetTasks(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
	if strings.Contains(capturedSQL, "WHERE") {
		t.Fatalf("expected no WHERE clause with no filters, got SQL: %s", capturedSQL)
	}
	if len(capturedArgs) != 0 {
		t.Fatalf("expected no args with no filters, got %v", capturedArgs)
	}
}

func TestStatisticsRepository_GetTasks_UserIDOnly(t *testing.T) {
	var capturedSQL string
	var capturedArgs []any

	pool := &mockPool{
		queryFn: func(_ context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
			capturedSQL, capturedArgs = sql, args
			return &mockRows{scanFns: []func(dest ...any) error{taskScanFn(1)}}, nil
		},
	}
	repo := NewStatisticsRepository(pool)

	userID := 5
	tasks, err := repo.GetTasks(context.Background(), &userID, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if !strings.Contains(capturedSQL, "author_user_id=$1") {
		t.Fatalf("expected author_user_id filter, got SQL: %s", capturedSQL)
	}
	if len(capturedArgs) != 1 {
		t.Fatalf("expected 1 arg, got %v", capturedArgs)
	}
}

func TestStatisticsRepository_GetTasks_AllFilters(t *testing.T) {
	var capturedSQL string
	var capturedArgs []any

	pool := &mockPool{
		queryFn: func(_ context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
			capturedSQL, capturedArgs = sql, args
			return &mockRows{}, nil
		},
	}
	repo := NewStatisticsRepository(pool)

	userID := 5
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	_, err := repo.GetTasks(context.Background(), &userID, &from, &to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedSQL, "author_user_id=$1") {
		t.Fatalf("expected author_user_id=$1, got SQL: %s", capturedSQL)
	}
	if !strings.Contains(capturedSQL, "created_at>=$2") {
		t.Fatalf("expected created_at>=$2, got SQL: %s", capturedSQL)
	}
	if !strings.Contains(capturedSQL, "created_at<$3") {
		t.Fatalf("expected created_at<$3, got SQL: %s", capturedSQL)
	}
	if len(capturedArgs) != 3 {
		t.Fatalf("expected 3 args (userID, from, to), got %d: %v", len(capturedArgs), capturedArgs)
	}
}

func TestStatisticsRepository_GetTasks_QueryError(t *testing.T) {
	pool := &mockPool{
		queryFn: func(_ context.Context, _ string, _ ...any) (core_postgres_pool.Rows, error) {
			return nil, errors.New("connection reset")
		},
	}
	repo := NewStatisticsRepository(pool)

	_, err := repo.GetTasks(context.Background(), nil, nil, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}
