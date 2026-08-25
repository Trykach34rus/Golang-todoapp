package users_postgres_repository

import (
	"context"
	"errors"
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
}

func (r *mockRows) Close()     {}
func (r *mockRows) Err() error { return nil }
func (r *mockRows) Next() bool {
	r.idx++
	return r.idx <= len(r.scanFns)
}
func (r *mockRows) Scan(dest ...any) error { return r.scanFns[r.idx-1](dest...) }

type mockCommandTag struct{ rowsAffected int64 }

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
func (p *mockPool) Close()                   {}
func (p *mockPool) OpTimeout() time.Duration { return time.Second }

func strPtr(s string) *string { return &s }

// --- CreateUser -----------------------------------------------------------

func TestUserRepository_CreateUser_Success(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, args ...any) core_postgres_pool.Row {
			if len(args) != 2 {
				t.Fatalf("expected 2 args (full_name, phone_number), got %d", len(args))
			}
			return mockRow{scanFn: func(dest ...any) error {
				*dest[0].(*int) = 1
				*dest[1].(*int) = 1
				*dest[2].(*string) = "Ivan Ivanov"
				*dest[3].(**string) = strPtr("+79876543210")
				return nil
			}}
		},
	}
	repo := NewUsersRepository(pool)

	user := domain.NewUserUninitialized("Ivan Ivanov", strPtr("+79876543210"))
	result, err := repo.CreateUser(context.Background(), user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 || result.FullName != "Ivan Ivanov" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestUserRepository_CreateUser_ScanError(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, _ ...any) core_postgres_pool.Row {
			return mockRow{scanFn: func(dest ...any) error {
				return errors.New("unique constraint violated")
			}}
		},
	}
	repo := NewUsersRepository(pool)

	_, err := repo.CreateUser(context.Background(), domain.NewUserUninitialized("Ivan Ivanov", nil))
	if err == nil {
		t.Fatalf("expected error")
	}
}

// --- GetUser -----------------------------------------------------------

func TestUserRepository_GetUser_Success(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, args ...any) core_postgres_pool.Row {
			if len(args) != 1 || args[0].(int) != 4 {
				t.Fatalf("expected id=4 as query arg, got %v", args)
			}
			return mockRow{scanFn: func(dest ...any) error {
				*dest[0].(*int) = 4
				*dest[1].(*int) = 1
				*dest[2].(*string) = "Ivan Ivanov"
				*dest[3].(**string) = nil
				return nil
			}}
		},
	}
	repo := NewUsersRepository(pool)

	user, err := repo.GetUser(context.Background(), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 4 || user.FullName != "Ivan Ivanov" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserRepository_GetUser_NotFound(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, _ ...any) core_postgres_pool.Row {
			return mockRow{scanFn: func(dest ...any) error {
				return core_postgres_pool.ErrNoRows
			}}
		},
	}
	repo := NewUsersRepository(pool)

	_, err := repo.GetUser(context.Background(), 999)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

// --- GetUsers -----------------------------------------------------------

func TestUserRepository_GetUsers_ReturnsAllRows(t *testing.T) {
	var capturedArgs []any
	pool := &mockPool{
		queryFn: func(_ context.Context, _ string, args ...any) (core_postgres_pool.Rows, error) {
			capturedArgs = args
			return &mockRows{scanFns: []func(dest ...any) error{
				func(dest ...any) error {
					*dest[0].(*int) = 1
					*dest[1].(*int) = 1
					*dest[2].(*string) = "User One"
					*dest[3].(**string) = nil
					return nil
				},
				func(dest ...any) error {
					*dest[0].(*int) = 2
					*dest[1].(*int) = 1
					*dest[2].(*string) = "User Two"
					*dest[3].(**string) = nil
					return nil
				},
			}}, nil
		},
	}
	repo := NewUsersRepository(pool)

	limit, offset := 10, 0
	users, err := repo.GetUsers(context.Background(), &limit, &offset)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if len(capturedArgs) != 2 {
		t.Fatalf("expected 2 args (limit, offset), got %d", len(capturedArgs))
	}
}

func TestUserRepository_GetUsers_QueryError(t *testing.T) {
	pool := &mockPool{
		queryFn: func(_ context.Context, _ string, _ ...any) (core_postgres_pool.Rows, error) {
			return nil, errors.New("connection reset")
		},
	}
	repo := NewUsersRepository(pool)

	limit, offset := 10, 0
	_, err := repo.GetUsers(context.Background(), &limit, &offset)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// --- DeleteUser -----------------------------------------------------------

func TestUserRepository_DeleteUser_Success(t *testing.T) {
	pool := &mockPool{
		execFn: func(_ context.Context, _ string, _ ...any) (core_postgres_pool.CommandTag, error) {
			return mockCommandTag{rowsAffected: 1}, nil
		},
	}
	repo := NewUsersRepository(pool)

	if err := repo.DeleteUser(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserRepository_DeleteUser_NotFound(t *testing.T) {
	pool := &mockPool{
		execFn: func(_ context.Context, _ string, _ ...any) (core_postgres_pool.CommandTag, error) {
			return mockCommandTag{rowsAffected: 0}, nil
		},
	}
	repo := NewUsersRepository(pool)

	err := repo.DeleteUser(context.Background(), 999)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when 0 rows affected, got: %v", err)
	}
}

// --- PatchUser (optimistic locking) -----------------------------------------

func TestUserRepository_PatchUser_Success(t *testing.T) {
	var capturedArgs []any
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, args ...any) core_postgres_pool.Row {
			capturedArgs = args
			return mockRow{scanFn: func(dest ...any) error {
				*dest[0].(*int) = 1
				*dest[1].(*int) = 2
				*dest[2].(*string) = "New Name"
				*dest[3].(**string) = nil
				return nil
			}}
		},
	}
	repo := NewUsersRepository(pool)

	user := domain.User{ID: 1, Version: 1, FullName: "New Name"}
	result, err := repo.PatchUser(context.Background(), 1, user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != 2 {
		t.Fatalf("expected version bumped to 2, got %d", result.Version)
	}
	if capturedArgs[len(capturedArgs)-2].(int) != 1 {
		t.Fatalf("expected id=1 in args, got %v", capturedArgs)
	}
	if capturedArgs[len(capturedArgs)-1].(int) != 1 {
		t.Fatalf("expected pre-patch version=1 in args, got %v", capturedArgs)
	}
}

func TestUserRepository_PatchUser_VersionConflict(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(_ context.Context, _ string, _ ...any) core_postgres_pool.Row {
			return mockRow{scanFn: func(dest ...any) error {
				return core_postgres_pool.ErrNoRows
			}}
		},
	}
	repo := NewUsersRepository(pool)

	user := domain.User{ID: 1, Version: 1, FullName: "New Name"}
	_, err := repo.PatchUser(context.Background(), 1, user)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict on optimistic lock failure, got: %v", err)
	}
}
