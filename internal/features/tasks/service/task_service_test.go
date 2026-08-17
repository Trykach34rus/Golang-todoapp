package task_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
)


type mockTaskRepository struct {
	createTaskFn func(ctx context.Context, task domain.Task) (domain.Task, error)
	getTaskFn    func(ctx context.Context, id int) (domain.Task, error)
	getTasksFn   func(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error)
	deleteTaskFn func(ctx context.Context, id int) error
	patchTaskFn  func(ctx context.Context, id int, task domain.Task) (domain.Task, error)
}

func (m *mockTaskRepository) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	return m.createTaskFn(ctx, task)
}

func (m *mockTaskRepository) GetTasks(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
	return m.getTasksFn(ctx, userID, limit, offset)
}

func (m *mockTaskRepository) GetTask(ctx context.Context, id int) (domain.Task, error) {
	return m.getTaskFn(ctx, id)
}

func (m *mockTaskRepository) DeleteTask(ctx context.Context, id int) error {
	return m.deleteTaskFn(ctx, id)
}

func (m *mockTaskRepository) PatchTask(ctx context.Context, id int, task domain.Task) (domain.Task, error) {
	return m.patchTaskFn(ctx, id, task)
}

func TestTaskService_CreateTask(t *testing.T) {
	t.Run("valid task is passed to repository", func(t *testing.T) {
		var receivedTask domain.Task
		repo := &mockTaskRepository{
			createTaskFn: func(_ context.Context, task domain.Task) (domain.Task, error) {
				receivedTask = task
				task.ID = 1
				return task, nil
			},
		}
		svc := NewTaskService(repo)

		task := domain.NewTaskUninitialized("Buy milk", nil, 1)

		result, err := svc.CreateTask(context.Background(), task)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != 1 {
			t.Fatalf("expected ID 1, got %d", result.ID)
		}
		if receivedTask.Title != "Buy milk" {
			t.Fatalf("expected repository to receive title 'Buy milk', got %q", receivedTask.Title)
		}
	})

	t.Run("invalid task never reaches repository", func(t *testing.T) {
		called := false
		repo := &mockTaskRepository{
			createTaskFn: func(_ context.Context, task domain.Task) (domain.Task, error) {
				called = true
				return task, nil
			},
		}
		svc := NewTaskService(repo)

		task := domain.NewTaskUninitialized("", nil, 1) // empty title -> invalid

		_, err := svc.CreateTask(context.Background(), task)
		if err == nil {
			t.Fatalf("expected validation error, got nil")
		}
		if called {
			t.Fatalf("expected repository NOT to be called for invalid task")
		}
	})

	t.Run("repository error is wrapped and returned", func(t *testing.T) {
		repoErr := errors.New("db is down")
		repo := &mockTaskRepository{
			createTaskFn: func(_ context.Context, task domain.Task) (domain.Task, error) {
				return domain.Task{}, repoErr
			},
		}
		svc := NewTaskService(repo)

		task := domain.NewTaskUninitialized("Buy milk", nil, 1)

		_, err := svc.CreateTask(context.Background(), task)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected wrapped repoErr, got: %v", err)
		}
	})
}

func TestTaskService_GetTasks(t *testing.T) {
	t.Run("negative limit is rejected before hitting repository", func(t *testing.T) {
		called := false
		repo := &mockTaskRepository{
			getTasksFn: func(_ context.Context, _ *int, _ *int, _ *int) ([]domain.Task, error) {
				called = true
				return nil, nil
			},
		}
		svc := NewTaskService(repo)

		negativeLimit := -1
		_, err := svc.GetTasks(context.Background(), nil, &negativeLimit, nil)
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
		repo := &mockTaskRepository{
			getTasksFn: func(_ context.Context, _ *int, _ *int, _ *int) ([]domain.Task, error) {
				called = true
				return nil, nil
			},
		}
		svc := NewTaskService(repo)

		negativeOffset := -1
		_, err := svc.GetTasks(context.Background(), nil, nil, &negativeOffset)
		if err == nil {
			t.Fatalf("expected error for negative offset")
		}
		if called {
			t.Fatalf("expected repository NOT to be called")
		}
	})

	t.Run("valid params delegate to repository", func(t *testing.T) {
		want := []domain.Task{{ID: 1, Title: "Buy milk"}}
		repo := &mockTaskRepository{
			getTasksFn: func(_ context.Context, _ *int, _ *int, _ *int) ([]domain.Task, error) {
				return want, nil
			},
		}
		svc := NewTaskService(repo)

		got, err := svc.GetTasks(context.Background(), nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].ID != 1 {
			t.Fatalf("expected tasks matching repository result, got: %v", got)
		}
	})
}

func TestTaskService_PatchTask(t *testing.T) {
	now := time.Now()

	t.Run("applies patch and persists via repository", func(t *testing.T) {
		existing := domain.Task{
			ID:        1,
			Title:     "Old title",
			CreatedAt: now,
		}

		var patchedReceived domain.Task
		repo := &mockTaskRepository{
			getTaskFn: func(_ context.Context, id int) (domain.Task, error) {
				return existing, nil
			},
			patchTaskFn: func(_ context.Context, id int, task domain.Task) (domain.Task, error) {
				patchedReceived = task
				return task, nil
			},
		}
		svc := NewTaskService(repo)

		newTitle := "New title"
		patch := domain.NewTaskPatch(
			domain.Nullable[string]{Value: &newTitle, Set: true},
			domain.Nullable[string]{},
			domain.Nullable[bool]{},
		)

		result, err := svc.PatchTask(context.Background(), 1, patch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Title != "New title" {
			t.Fatalf("expected title 'New title', got %q", result.Title)
		}
		if patchedReceived.Title != "New title" {
			t.Fatalf("expected repository to receive patched title")
		}
	})

	t.Run("task not found propagates error", func(t *testing.T) {
		repo := &mockTaskRepository{
			getTaskFn: func(_ context.Context, id int) (domain.Task, error) {
				return domain.Task{}, core_errors.ErrNotFound
			},
		}
		svc := NewTaskService(repo)

		newTitle := "New title"
		patch := domain.NewTaskPatch(
			domain.Nullable[string]{Value: &newTitle, Set: true},
			domain.Nullable[string]{},
			domain.Nullable[bool]{},
		)

		_, err := svc.PatchTask(context.Background(), 999, patch)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("invalid patch never reaches repository PatchTask", func(t *testing.T) {
		existing := domain.Task{ID: 1, Title: "Old title", CreatedAt: now}
		patchCalled := false

		repo := &mockTaskRepository{
			getTaskFn: func(_ context.Context, id int) (domain.Task, error) {
				return existing, nil
			},
			patchTaskFn: func(_ context.Context, id int, task domain.Task) (domain.Task, error) {
				patchCalled = true
				return task, nil
			},
		}
		svc := NewTaskService(repo)

		// Title patched to nil is invalid.
		patch := domain.NewTaskPatch(
			domain.Nullable[string]{Value: nil, Set: true},
			domain.Nullable[string]{},
			domain.Nullable[bool]{},
		)

		_, err := svc.PatchTask(context.Background(), 1, patch)
		if err == nil {
			t.Fatalf("expected error for invalid patch")
		}
		if patchCalled {
			t.Fatalf("expected repository.PatchTask NOT to be called")
		}
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		called := false
		repo := &mockTaskRepository{
			deleteTaskFn: func(_ context.Context, id int) error {
				called = true
				if id != 5 {
					t.Fatalf("expected id 5, got %d", id)
				}
				return nil
			},
		}
		svc := NewTaskService(repo)

		if err := svc.DeleteTask(context.Background(), 5); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("expected repository.DeleteTask to be called")
		}
	})

	t.Run("not found error is propagated", func(t *testing.T) {
		repo := &mockTaskRepository{
			deleteTaskFn: func(_ context.Context, id int) error {
				return core_errors.ErrNotFound
			},
		}
		svc := NewTaskService(repo)

		err := svc.DeleteTask(context.Background(), 5)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}
