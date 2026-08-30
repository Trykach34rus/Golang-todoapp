package cached_repository

import (
	"context"
	"fmt"
)

func (r *CachedRepository) DeleteTask(
    ctx context.Context,
    id int,
) error {

    task, ok := r.getTaskFromCache(ctx, id)

    if !ok {
        var err error

        task, err = r.mainRepository.GetTask(ctx, id)
        if err != nil {
            return fmt.Errorf("get task info: %w", err)
        }
    }

    r.invalidateTasks(
        ctx,
        task.AuthorUserID,
        &id,
    )

    return r.mainRepository.DeleteTask(ctx, id)
}