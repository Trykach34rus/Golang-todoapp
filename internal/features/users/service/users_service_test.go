package users_service

import (
	"context"
	"errors"
	"testing"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
)

type mockUsersRepository struct {
	createUserFn func(ctx context.Context, user domain.User) (domain.User, error)
	getUserFn    func(ctx context.Context, id int) (domain.User, error)
	getUsersFn   func(ctx context.Context, limit *int, offset *int) ([]domain.User, error)
	deleteUserFn func(ctx context.Context, id int) error
	patchUserFn  func(ctx context.Context, id int, user domain.User) (domain.User, error)
}

func (m *mockUsersRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return m.createUserFn(ctx, user)
}

func (m *mockUsersRepository) GetUsers(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
	return m.getUsersFn(ctx, limit, offset)
}

func (m *mockUsersRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	return m.getUserFn(ctx, id)
}

func (m *mockUsersRepository) DeleteUser(ctx context.Context, id int) error {
	return m.deleteUserFn(ctx, id)
}

func (m *mockUsersRepository) PatchUser(ctx context.Context, id int, user domain.User) (domain.User, error) {
	return m.patchUserFn(ctx, id, user)
}

func strPtr(s string) *string { return &s }

func TestUsersService_CreateUser(t *testing.T) {
	t.Run("valid user is passed to repository", func(t *testing.T) {
		var receivedUser domain.User
		repo := &mockUsersRepository{
			createUserFn: func(_ context.Context, user domain.User) (domain.User, error) {
				receivedUser = user
				user.ID = 1
				return user, nil
			},
		}
		svc := NewUsersService(repo)

		user := domain.NewUserUninitialized("Ivan Ivanov", nil)

		result, err := svc.CreateUser(context.Background(), user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != 1 {
			t.Fatalf("expected ID 1, got %d", result.ID)
		}
		if receivedUser.FullName != "Ivan Ivanov" {
			t.Fatalf("expected repository to receive full name 'Ivan Ivanov', got %q", receivedUser.FullName)
		}
	})

	t.Run("invalid user never reaches repository", func(t *testing.T) {
		called := false
		repo := &mockUsersRepository{
			createUserFn: func(_ context.Context, user domain.User) (domain.User, error) {
				called = true
				return user, nil
			},
		}
		svc := NewUsersService(repo)

		user := domain.NewUserUninitialized("Iv", nil) // too short -> invalid

		_, err := svc.CreateUser(context.Background(), user)
		if err == nil {
			t.Fatalf("expected validation error, got nil")
		}
		if called {
			t.Fatalf("expected repository NOT to be called for invalid user")
		}
	})

	t.Run("invalid phone number is rejected before hitting repository", func(t *testing.T) {
		called := false
		repo := &mockUsersRepository{
			createUserFn: func(_ context.Context, user domain.User) (domain.User, error) {
				called = true
				return user, nil
			},
		}
		svc := NewUsersService(repo)

		user := domain.NewUserUninitialized("Ivan Ivanov", strPtr("not-a-phone"))

		_, err := svc.CreateUser(context.Background(), user)
		if err == nil {
			t.Fatalf("expected validation error for invalid phone number")
		}
		if called {
			t.Fatalf("expected repository NOT to be called")
		}
	})

	t.Run("repository error is wrapped and returned", func(t *testing.T) {
		repoErr := errors.New("db is down")
		repo := &mockUsersRepository{
			createUserFn: func(_ context.Context, user domain.User) (domain.User, error) {
				return domain.User{}, repoErr
			},
		}
		svc := NewUsersService(repo)

		user := domain.NewUserUninitialized("Ivan Ivanov", nil)

		_, err := svc.CreateUser(context.Background(), user)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestUsersService_GetUsers(t *testing.T) {
	t.Run("negative limit is rejected before hitting repository", func(t *testing.T) {
		called := false
		repo := &mockUsersRepository{
			getUsersFn: func(_ context.Context, _ *int, _ *int) ([]domain.User, error) {
				called = true
				return nil, nil
			},
		}
		svc := NewUsersService(repo)

		negativeLimit := -1
		_, err := svc.GetUsers(context.Background(), &negativeLimit, nil)
		if err == nil {
			t.Fatalf("expected error for negative limit")
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument, got: %v", err)
		}
		if called {
			t.Fatalf("expected repository NOT to be called")
		}
	})

	t.Run("negative offset is rejected before hitting repository", func(t *testing.T) {
		called := false
		repo := &mockUsersRepository{
			getUsersFn: func(_ context.Context, _ *int, _ *int) ([]domain.User, error) {
				called = true
				return nil, nil
			},
		}
		svc := NewUsersService(repo)

		negativeOffset := -1
		_, err := svc.GetUsers(context.Background(), nil, &negativeOffset)
		if err == nil {
			t.Fatalf("expected error for negative offset")
		}
		if called {
			t.Fatalf("expected repository NOT to be called")
		}
	})

	t.Run("valid params delegate to repository", func(t *testing.T) {
		want := []domain.User{{ID: 1, FullName: "Ivan Ivanov"}}
		repo := &mockUsersRepository{
			getUsersFn: func(_ context.Context, _ *int, _ *int) ([]domain.User, error) {
				return want, nil
			},
		}
		svc := NewUsersService(repo)

		got, err := svc.GetUsers(context.Background(), nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].ID != 1 {
			t.Fatalf("expected users matching repository result, got: %v", got)
		}
	})
}

func TestUsersService_GetUser(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		repo := &mockUsersRepository{
			getUserFn: func(_ context.Context, id int) (domain.User, error) {
				return domain.User{ID: id, FullName: "Ivan Ivanov"}, nil
			},
		}
		svc := NewUsersService(repo)

		user, err := svc.GetUser(context.Background(), 7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != 7 {
			t.Fatalf("expected ID 7, got %d", user.ID)
		}
	})

	t.Run("not found error is propagated", func(t *testing.T) {
		repo := &mockUsersRepository{
			getUserFn: func(_ context.Context, id int) (domain.User, error) {
				return domain.User{}, core_errors.ErrNotFound
			},
		}
		svc := NewUsersService(repo)

		_, err := svc.GetUser(context.Background(), 999)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestUsersService_PatchUser(t *testing.T) {
	t.Run("applies patch and persists via repository", func(t *testing.T) {
		existing := domain.User{ID: 1, FullName: "Old Name"}

		var patchedReceived domain.User
		repo := &mockUsersRepository{
			getUserFn: func(_ context.Context, id int) (domain.User, error) {
				return existing, nil
			},
			patchUserFn: func(_ context.Context, id int, user domain.User) (domain.User, error) {
				patchedReceived = user
				return user, nil
			},
		}
		svc := NewUsersService(repo)

		newName := "New Name"
		patch := domain.NewUserPatch(
			domain.Nullable[string]{Value: &newName, Set: true},
			domain.Nullable[string]{},
		)

		result, err := svc.PatchUser(context.Background(), 1, patch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.FullName != "New Name" {
			t.Fatalf("expected full name 'New Name', got %q", result.FullName)
		}
		if patchedReceived.FullName != "New Name" {
			t.Fatalf("expected repository to receive patched full name")
		}
	})

	t.Run("user not found propagates error", func(t *testing.T) {
		repo := &mockUsersRepository{
			getUserFn: func(_ context.Context, id int) (domain.User, error) {
				return domain.User{}, core_errors.ErrNotFound
			},
		}
		svc := NewUsersService(repo)

		newName := "New Name"
		patch := domain.NewUserPatch(
			domain.Nullable[string]{Value: &newName, Set: true},
			domain.Nullable[string]{},
		)

		_, err := svc.PatchUser(context.Background(), 999, patch)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("invalid patch never reaches repository PatchUser", func(t *testing.T) {
		existing := domain.User{ID: 1, FullName: "Old Name"}
		patchCalled := false

		repo := &mockUsersRepository{
			getUserFn: func(_ context.Context, id int) (domain.User, error) {
				return existing, nil
			},
			patchUserFn: func(_ context.Context, id int, user domain.User) (domain.User, error) {
				patchCalled = true
				return user, nil
			},
		}
		svc := NewUsersService(repo)

		// FullName patched to nil is invalid.
		patch := domain.NewUserPatch(
			domain.Nullable[string]{Value: nil, Set: true},
			domain.Nullable[string]{},
		)

		_, err := svc.PatchUser(context.Background(), 1, patch)
		if err == nil {
			t.Fatalf("expected error for invalid patch")
		}
		if patchCalled {
			t.Fatalf("expected repository.PatchUser NOT to be called")
		}
	})

	t.Run("patch with invalid phone number is rejected after applying", func(t *testing.T) {
		existing := domain.User{ID: 1, FullName: "Old Name"}
		patchCalled := false

		repo := &mockUsersRepository{
			getUserFn: func(_ context.Context, id int) (domain.User, error) {
				return existing, nil
			},
			patchUserFn: func(_ context.Context, id int, user domain.User) (domain.User, error) {
				patchCalled = true
				return user, nil
			},
		}
		svc := NewUsersService(repo)

		patch := domain.NewUserPatch(
			domain.Nullable[string]{},
			domain.Nullable[string]{Value: strPtr("bad-phone"), Set: true},
		)

		_, err := svc.PatchUser(context.Background(), 1, patch)
		if err == nil {
			t.Fatalf("expected error for invalid phone number")
		}
		if patchCalled {
			t.Fatalf("expected repository.PatchUser NOT to be called")
		}
	})
}

func TestUsersService_DeleteUser(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		called := false
		repo := &mockUsersRepository{
			deleteUserFn: func(_ context.Context, id int) error {
				called = true
				if id != 5 {
					t.Fatalf("expected id 5, got %d", id)
				}
				return nil
			},
		}
		svc := NewUsersService(repo)

		if err := svc.DeleteUser(context.Background(), 5); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("expected repository.DeleteUser to be called")
		}
	})

	t.Run("not found error is propagated", func(t *testing.T) {
		repo := &mockUsersRepository{
			deleteUserFn: func(_ context.Context, id int) error {
				return core_errors.ErrNotFound
			},
		}
		svc := NewUsersService(repo)

		err := svc.DeleteUser(context.Background(), 5)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}
