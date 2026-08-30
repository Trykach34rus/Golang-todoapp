package cached_repository

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
)

type TaskModel struct {
	ID          int    `json:"id"`
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description *string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
	AuthorUserID int `json:"author_user_id"`
}

func domainToModel(task domain.Task) TaskModel {
	return TaskModel{
		ID:           task.ID,
		Version:      task.Version,
		Title:        task.Title,
		Description:  task.Description,
		Completed:    task.Completed,
		CreatedAt:    task.CreatedAt,
		CompletedAt:  task.CompletedAt,
		AuthorUserID: task.AuthorUserID,
	}
}

func modelToDomain(model TaskModel) domain.Task {
	return domain.NewTask(
		model.ID,
		model.Version,
		model.Title,
		model.Description,
		model.Completed,
		model.CreatedAt,
		model.CompletedAt,
		model.AuthorUserID,
	)
}

func taskKey(id int) string {
	return fmt.Sprintf("task:%d", id)
}

func (m *TaskModel)Serialize()([]byte,error)  {
	bytes,err := json.Marshal(m)
	if err != nil {
		return nil,fmt.Errorf("serialize task: %w",err)
	}

	return bytes,nil
}

func (m *TaskModel)Deserialize(bytes []byte) error  {
	if err := json.Unmarshal(bytes,m); err != nil {
		return fmt.Errorf("deserialize task: %w",err)
	}

	return nil
}

type TaskListModel []TaskModel

func domainsToModel(tasks []domain.Task) TaskListModel {
	tasksModels := make([]TaskModel, len(tasks))

	for i, task := range tasks {
		tasksModels[i] = domainToModel(task)
	}

	return tasksModels
}

func modelToDomains(list TaskListModel) []domain.Task  {
	tasks := make([]domain.Task, len(list))

	for i,model := range list {
		tasks[i] = modelToDomain(model)
	}

	return tasks
}

func tasksListKey(userID *int) string {
	if userID == nil {
		return "tasks:all"
	}

	return fmt.Sprintf("tasks:%d", *userID)
}

func tasksListField(limit *int, offset *int) string {
	ptrStr := func(v *int) string {
		if v == nil {
			return "nil"
		}

		return strconv.Itoa(*v)
	}

	return fmt.Sprintf("%s:%s", ptrStr(limit), ptrStr(offset))
}

func (m *TaskListModel) Serialize() ([]byte, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("serialize task list: %w", err)
	}

	return bytes, err
}

func (m *TaskListModel) Deserialize(bytes []byte) error {
	if err := json.Unmarshal(bytes, m); err != nil {
		return fmt.Errorf("deserialize task list: %w", err)
	}

	return nil
}