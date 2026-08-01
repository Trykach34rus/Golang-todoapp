package task_postgres_repository

import core_postgres_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/postges/pool"

type TaskRepository struct {
	pool core_postgres_pool.Pool
}

func NewTaskRepository(
	pool core_postgres_pool.Pool,
) *TaskRepository {
	return &TaskRepository{
		pool: pool,
	}
}