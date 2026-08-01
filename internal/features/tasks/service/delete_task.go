package task_service

import (
	"context"
	"fmt"
)

func (s *TaskService) DeleteTask(
	ctx context.Context,
	id int,
) error {
	if err := s.tasksRepository.DeleteTask(ctx,id); err != nil {
		return fmt.Errorf("delete task: %w",err)
	}
	return nil
}