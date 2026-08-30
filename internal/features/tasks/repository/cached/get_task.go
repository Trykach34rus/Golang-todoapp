package cached_repository

import (
	"context"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
)

func (r *CachedRepository) GetTask(
	ctx context.Context,
	id int,
) (domain.Task, error) {

	task, ok := r.getTaskFromCache(ctx, id)
	if ok {
		return task, nil
	}

	task, err := r.mainRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}

	r.cacheTask(ctx, task)

	return task, nil
}