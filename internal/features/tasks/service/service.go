package task_service

import (
	"context"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
)

type TaskService struct {
	tasksRepository TaskRepository
}

type TaskRepository interface {
	CreateTask(
		ctx context.Context,
		task domain.Task,
	) (domain.Task, error)
	GetTasks(
		ctx context.Context,
		userID *int,
		limit *int,
		offset *int,		
	)([]domain.Task,error)
	GetTask(
		ctx context.Context,
		id int,
	)(domain.Task,error)
	DeleteTask(
		ctx context.Context,
		id int,
	) error
	PatchTask(
		ctx context.Context,
		id int,
		task domain.Task,
	) (domain.Task,error)

}

func NewTaskService(
	tasksRepository TaskRepository,
) *TaskService {
	return &TaskService{
		tasksRepository: tasksRepository,
	}
}