package cached_repository

import (
	"context"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
)

func (r *CachedRepository) CreateTask(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {

	// 1. Создаём задачу в PostgreSQL
	task, err := r.mainRepository.CreateTask(ctx, task)
	if err != nil {
		return domain.Task{}, err
	}

	// 2. Списки задач стали неактуальными
	r.invalidateTasks(
		ctx,
		task.AuthorUserID,
		nil,
	)

	return task, nil
}