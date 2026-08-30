package cached_repository

import (
	"context"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
)

func (r *CachedRepository) PatchTask(
	ctx context.Context,
	id int,
	task domain.Task,
) (domain.Task, error) {

	task, err := r.mainRepository.PatchTask(
		ctx,
		id,
		task,
	)
	if err != nil {
		return domain.Task{}, err
	}

	r.cacheTask(ctx, task)

	r.invalidateTasks(
		ctx,
		task.AuthorUserID,
		nil,
	)

	return task, nil
}