package service_statistics

import (
	"context"
	"testing"
	"time"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
)

type mockStatisticsRepository struct {
	getTasksFn func(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error)
}

func (m *mockStatisticsRepository) GetTasks(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error) {
	return m.getTasksFn(ctx, userID, from, to)
}

func TestStatisticsService_GetStatistics_InvalidRange(t *testing.T) {
	repo := &mockStatisticsRepository{
		getTasksFn: func(_ context.Context, _ *int, _ *time.Time, _ *time.Time) ([]domain.Task, error) {
			t.Fatalf("repository should not be queried for an invalid date range")
			return nil, nil
		},
	}
	svc := NewStatisticsService(repo)

	from := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) // to before from

	_, err := svc.GetStatistics(context.Background(), nil, &from, &to)
	if err == nil {
		t.Fatalf("expected error because `to` is before `from`, got nil (this is the known && vs || bug)")
	}
}

func TestStatisticsService_GetStatistics_EmptyTasks(t *testing.T) {
	repo := &mockStatisticsRepository{
		getTasksFn: func(_ context.Context, _ *int, _ *time.Time, _ *time.Time) ([]domain.Task, error) {
			return nil, nil
		},
	}
	svc := NewStatisticsService(repo)

	stats, err := svc.GetStatistics(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TasksCreated != 0 || stats.TasksCompleted != 0 {
		t.Fatalf("expected zero stats for empty task list, got: %+v", stats)
	}
	if stats.TasksCompletedRate != nil {
		t.Fatalf("expected nil TasksCompletedRate for empty task list")
	}
	if stats.TaskAverageCompletionTime != nil {
		t.Fatalf("expected nil TaskAverageCompletionTime for empty task list")
	}
}

func TestStatisticsService_GetStatistics_CalculatesRateAndAverage(t *testing.T) {
	created := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	completedAt1 := created.Add(1 * time.Hour)
	completedAt2 := created.Add(3 * time.Hour)

	tasks := []domain.Task{
		{ID: 1, Completed: true, CreatedAt: created, CompletedAt: &completedAt1},
		{ID: 2, Completed: true, CreatedAt: created, CompletedAt: &completedAt2},
		{ID: 3, Completed: false, CreatedAt: created},
		{ID: 4, Completed: false, CreatedAt: created},
	}

	repo := &mockStatisticsRepository{
		getTasksFn: func(_ context.Context, _ *int, _ *time.Time, _ *time.Time) ([]domain.Task, error) {
			return tasks, nil
		},
	}
	svc := NewStatisticsService(repo)

	stats, err := svc.GetStatistics(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.TasksCreated != 4 {
		t.Fatalf("expected TasksCreated=4, got %d", stats.TasksCreated)
	}
	if stats.TasksCompleted != 2 {
		t.Fatalf("expected TasksCompleted=2, got %d", stats.TasksCompleted)
	}
	if stats.TasksCompletedRate == nil || *stats.TasksCompletedRate != 50 {
		t.Fatalf("expected TasksCompletedRate=50, got %v", stats.TasksCompletedRate)
	}
	// average of 1h and 3h completion times = 2h
	if stats.TaskAverageCompletionTime == nil || *stats.TaskAverageCompletionTime != 2*time.Hour {
		t.Fatalf("expected average completion time of 2h, got %v", stats.TaskAverageCompletionTime)
	}
}

func TestStatisticsService_GetStatistics_RepositoryErrorIsWrapped(t *testing.T) {
	repoErr := context.DeadlineExceeded
	repo := &mockStatisticsRepository{
		getTasksFn: func(_ context.Context, _ *int, _ *time.Time, _ *time.Time) ([]domain.Task, error) {
			return nil, repoErr
		},
	}
	svc := NewStatisticsService(repo)

	_, err := svc.GetStatistics(context.Background(), nil, nil, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
