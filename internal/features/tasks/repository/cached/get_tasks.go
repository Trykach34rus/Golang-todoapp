package cached_repository

import (
	"context"
	"errors"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_redis_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/redis/pool"
	"go.uber.org/zap"
)

func (r *CachedRepository) GetTasks(
	ctx context.Context,
	userID *int,
	limit *int,
	offset *int,
) ([]domain.Task, error) {

	log := core_logger.FromContext(ctx)

	key := tasksListKey(userID)
	field := tasksListField(limit, offset)

	bytes, err := r.pool.HGet(ctx, key, field).Bytes()

	if err == nil {
		var taskListModel TaskListModel

		if err := taskListModel.Deserialize(bytes); err != nil {
			log.Error("deserialize cached task list", zap.Error(err))
		} else {
			return modelToDomains(taskListModel), nil
		}
	} else if !errors.Is(err, core_redis_pool.NotFound) {
		log.Error("hget task list", zap.Error(err))
	}

	tasks, err := r.mainRepository.GetTasks(
		ctx,
		userID,
		limit,
		offset,
	)

	if err != nil {
		return nil, err
	}

	taskListModel := domainsToModel(tasks)

	bytes, err = taskListModel.Serialize()
	if err != nil {
		log.Error("serialize task list", zap.Error(err))
		return tasks, nil
	}

	if err := r.pool.HSet(
		ctx,
		key,
		field,
		bytes,
	).Err(); err != nil {
		log.Error("hset task list in cache", zap.Error(err))
	}

	return tasks, nil
}