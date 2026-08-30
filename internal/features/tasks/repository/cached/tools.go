package cached_repository

import (
	"context"
	"errors"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_redis_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/redis/pool"
	"go.uber.org/zap"
)

func (r *CachedRepository) getTaskFromCache(
	ctx context.Context,
	id int,
) (domain.Task,bool) {
	log := core_logger.FromContext(ctx)

	key := taskKey(id)

	bytes,err := r.pool.Get(ctx,key).Bytes()
	if err != nil {
		if !errors.Is(err,core_redis_pool.NotFound){
			log.Error("read from cache",zap.Error(err))
		}
		return domain.Task{},false
	}

	var taskModel TaskModel
	if err := taskModel.Deserialize(bytes);err != nil {
		log.Error("deserialize cached task",zap.Error(err))

		return domain.Task{},false
	}

	taskDomain := modelToDomain(taskModel)

	return taskDomain,true
}

func (r *CachedRepository)cacheTask(
	ctx context.Context,
	task domain.Task,
) {
	log := core_logger.FromContext(ctx)

	taskModel := domainToModel(task)
	bytes,err := taskModel.Serialize()
	if err != nil {
		log.Error("serialize task",zap.Error(err))
	}else {
		if err := r.pool.Set(
			ctx,
			taskKey(taskModel.ID),
			bytes,
			r.pool.TTL(),
		).Err(); err != nil {
			log.Error("set task in cache",zap.Error(err))
		}
	}
}

func (r *CachedRepository) invalidateTasks(
	ctx context.Context,
	userID int,
	taskID *int,
)  {
	log := core_logger.FromContext(ctx)

	invaldateKeys := []string{
		tasksListKey(nil),
		tasksListKey(&userID),
	}
	if taskID != nil {
		invaldateKeys = append(invaldateKeys, taskKey(*taskID))
	}

	if err := r.pool.Del(ctx,invaldateKeys...).Err();err !=nil{
		log.Error("invaldate cached tasks list",zap.Error(err))
	}
}