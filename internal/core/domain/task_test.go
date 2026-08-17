package domain

import (
	"errors"
	"testing"
	"time"

	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
)

func strPtr(s string) *string { return &s }

func TestTask_Validate(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(time.Hour)

	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{
			name: "valid uncompleted task",
			task: Task{
				Title:     "Buy milk",
				CreatedAt: now,
				Completed: false,
			},
			wantErr: false,
		},
		{
			name: "valid completed task",
			task: Task{
				Title:       "Buy milk",
				CreatedAt:   now,
				Completed:   true,
				CompletedAt: &completedAt,
			},
			wantErr: false,
		},
		{
			name: "empty title",
			task: Task{
				Title:     "",
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "title too long (101 chars)",
			task: Task{
				Title:     string(make([]rune, 101)),
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "empty description pointer set to empty string",
			task: Task{
				Title:       "Buy milk",
				Description: strPtr(""),
				CreatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "completed true but CompletedAt nil",
			task: Task{
				Title:     "Buy milk",
				CreatedAt: now,
				Completed: true,
			},
			wantErr: true,
		},
		{
			name: "CompletedAt before CreatedAt",
			task: Task{
				Title:       "Buy milk",
				CreatedAt:   now,
				Completed:   true,
				CompletedAt: &now,
			},
			wantErr: false, // equal, not before -- boundary check
		},
		{
			name: "CompletedAt strictly before CreatedAt",
			task: Task{
				Title:       "Buy milk",
				CreatedAt:   now,
				Completed:   true,
				CompletedAt: func() *time.Time { c := now.Add(-time.Hour); return &c }(),
			},
			wantErr: true,
		},
		{
			name: "completed false but CompletedAt set",
			task: Task{
				Title:       "Buy milk",
				CreatedAt:   now,
				Completed:   false,
				CompletedAt: &now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tt.wantErr && err != nil && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
			}
		})
	}
}

func TestTask_CompletionDuration(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(2 * time.Hour)

	t.Run("not completed returns nil", func(t *testing.T) {
		task := Task{Completed: false, CreatedAt: now}
		if d := task.CompletionDuration(); d != nil {
			t.Fatalf("expected nil, got %v", d)
		}
	})

	t.Run("completed but no CompletedAt returns nil", func(t *testing.T) {
		task := Task{Completed: true, CreatedAt: now, CompletedAt: nil}
		if d := task.CompletionDuration(); d != nil {
			t.Fatalf("expected nil, got %v", d)
		}
	})

	t.Run("completed with CompletedAt returns correct duration", func(t *testing.T) {
		task := Task{Completed: true, CreatedAt: now, CompletedAt: &completedAt}
		d := task.CompletionDuration()
		if d == nil {
			t.Fatalf("expected non-nil duration")
		}
		if *d != 2*time.Hour {
			t.Fatalf("expected 2h, got %v", *d)
		}
	})
}

func TestTask_ApplyPatch(t *testing.T) {
	now := time.Now()

	t.Run("patch title only", func(t *testing.T) {
		task := Task{Title: "old", CreatedAt: now}
		patch := NewTaskPatch(
			Nullable[string]{Value: strPtr("new title"), Set: true},
			Nullable[string]{},
			Nullable[bool]{},
		)

		if err := task.ApplyPatch(patch); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if task.Title != "new title" {
			t.Fatalf("expected title 'new title', got %q", task.Title)
		}
	})

	t.Run("patch title to nil is rejected", func(t *testing.T) {
		task := Task{Title: "old", CreatedAt: now}
		patch := NewTaskPatch(
			Nullable[string]{Value: nil, Set: true},
			Nullable[string]{},
			Nullable[bool]{},
		)

		if err := task.ApplyPatch(patch); err == nil {
			t.Fatalf("expected error when patching Title to NULL")
		}
	})

	t.Run("patch completed to true sets CompletedAt", func(t *testing.T) {
		task := Task{Title: "old", CreatedAt: now, Completed: false}
		completed := true
		patch := NewTaskPatch(
			Nullable[string]{},
			Nullable[string]{},
			Nullable[bool]{Value: &completed, Set: true},
		)

		if err := task.ApplyPatch(patch); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !task.Completed {
			t.Fatalf("expected task to be completed")
		}
		if task.CompletedAt == nil {
			t.Fatalf("expected CompletedAt to be set")
		}
	})

	t.Run("patch completed to false clears CompletedAt", func(t *testing.T) {
		completedAt := now.Add(time.Hour)
		task := Task{Title: "old", CreatedAt: now, Completed: true, CompletedAt: &completedAt}
		completed := false
		patch := NewTaskPatch(
			Nullable[string]{},
			Nullable[string]{},
			Nullable[bool]{Value: &completed, Set: true},
		)

		if err := task.ApplyPatch(patch); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if task.Completed {
			t.Fatalf("expected task to not be completed")
		}
		if task.CompletedAt != nil {
			t.Fatalf("expected CompletedAt to be nil")
		}
	})

	t.Run("patch description with invalid length is rejected", func(t *testing.T) {
		task := Task{Title: "old", CreatedAt: now}
		patch := NewTaskPatch(
			Nullable[string]{},
			Nullable[string]{Value: strPtr(""), Set: true},
			Nullable[bool]{},
		)

		if err := task.ApplyPatch(patch); err == nil {
			t.Fatalf("expected error for empty description")
		}
	})
}
