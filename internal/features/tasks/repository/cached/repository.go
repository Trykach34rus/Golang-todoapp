package cached_repository

import (
	core_redis_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/redis/pool"
	task_service "github.com/Trykach34rus/Golang-todoapp/internal/features/tasks/service"
)

type CachedRepository struct {
	pool core_redis_pool.Pool
	mainRepository task_service.TaskRepository
}

func NewCachedRepository(
	pool core_redis_pool.Pool,
	mainRepository task_service.TaskRepository,
) *CachedRepository {
	return &CachedRepository{
		pool: pool,
		mainRepository: mainRepository,
	}
}