package tasks_transport_http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_http_response "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/response"
	"go.uber.org/zap"
)


type mockTasksService struct {
	createTaskFn func(ctx context.Context, task domain.Task) (domain.Task, error)
	getTaskFn    func(ctx context.Context, id int) (domain.Task, error)
	getTasksFn   func(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error)
	deleteTaskFn func(ctx context.Context, id int) error
	patchTaskFn  func(ctx context.Context, id int, patch domain.TaskPatch) (domain.Task, error)
}

func (m *mockTasksService) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	return m.createTaskFn(ctx, task)
}

func (m *mockTasksService) GetTasks(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
	return m.getTasksFn(ctx, userID, limit, offset)
}

func (m *mockTasksService) GetTask(ctx context.Context, id int) (domain.Task, error) {
	return m.getTaskFn(ctx, id)
}

func (m *mockTasksService) DeleteTask(ctx context.Context, id int) error {
	return m.deleteTaskFn(ctx, id)
}

func (m *mockTasksService) PatchTask(ctx context.Context, id int, patch domain.TaskPatch) (domain.Task, error) {
	return m.patchTaskFn(ctx, id, patch)
}


func withLoggerContext(req *http.Request) *http.Request {
	log := &core_logger.Logger{Logger: zap.NewNop()}
	return req.WithContext(core_logger.ToContext(req.Context(), log))
}

func newJSONRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return withLoggerContext(req)
}


func TestTasksHandler_CreateTask_Valid(t *testing.T) {
	var receivedTask domain.Task
	svc := &mockTasksService{
		createTaskFn: func(_ context.Context, task domain.Task) (domain.Task, error) {
			receivedTask = task
			task.ID = 1
			return task, nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := newJSONRequest(http.MethodPost, "/tasks", `{"title":"Buy milk","author_user_id":1}`)
	rw := httptest.NewRecorder()

	h.CreateTask(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rw.Code, rw.Body.String())
	}
	if receivedTask.Title != "Buy milk" {
		t.Fatalf("expected service to receive title 'Buy milk', got %q", receivedTask.Title)
	}

	var resp CreateTaskResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response, got: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("expected ID 1 in response, got %d", resp.ID)
	}
}

func TestTasksHandler_CreateTask_InvalidBody(t *testing.T) {
	called := false
	svc := &mockTasksService{
		createTaskFn: func(_ context.Context, task domain.Task) (domain.Task, error) {
			called = true
			return task, nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := newJSONRequest(http.MethodPost, "/tasks", `{}`)
	rw := httptest.NewRecorder()

	h.CreateTask(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rw.Code)
	}
	if called {
		t.Fatalf("expected service NOT to be called for invalid body")
	}
}

func TestTasksHandler_CreateTask_ServiceNotFoundError(t *testing.T) {
	svc := &mockTasksService{
		createTaskFn: func(_ context.Context, task domain.Task) (domain.Task, error) {
			return domain.Task{}, core_errors.ErrNotFound
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := newJSONRequest(http.MethodPost, "/tasks", `{"title":"Buy milk","author_user_id":999}`)
	rw := httptest.NewRecorder()

	h.CreateTask(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 when author does not exist, got %d", rw.Code)
	}
}


func TestTasksHandler_GetTask_Valid(t *testing.T) {
	svc := &mockTasksService{
		getTaskFn: func(_ context.Context, id int) (domain.Task, error) {
			return domain.Task{ID: id, Title: "Buy milk"}, nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/tasks/7", nil))
	req.SetPathValue("id", "7")
	rw := httptest.NewRecorder()

	h.GetTask(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}

	var resp GetTaskResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if resp.ID != 7 {
		t.Fatalf("expected ID 7, got %d", resp.ID)
	}
}

func TestTasksHandler_GetTask_InvalidPathValue(t *testing.T) {
	called := false
	svc := &mockTasksService{
		getTaskFn: func(_ context.Context, id int) (domain.Task, error) {
			called = true
			return domain.Task{}, nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/tasks/abc", nil))
	req.SetPathValue("id", "abc")
	rw := httptest.NewRecorder()

	h.GetTask(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for non-integer id, got %d", rw.Code)
	}
	if called {
		t.Fatalf("expected service NOT to be called")
	}
}

func TestTasksHandler_GetTask_NotFound(t *testing.T) {
	svc := &mockTasksService{
		getTaskFn: func(_ context.Context, id int) (domain.Task, error) {
			return domain.Task{}, core_errors.ErrNotFound
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/tasks/999", nil))
	req.SetPathValue("id", "999")
	rw := httptest.NewRecorder()

	h.GetTask(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rw.Code)
	}
}


func TestTasksHandler_GetTasks_ParsesQueryParamsAndReturnsList(t *testing.T) {
	var receivedUserID, receivedLimit, receivedOffset *int
	svc := &mockTasksService{
		getTasksFn: func(_ context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
			receivedUserID, receivedLimit, receivedOffset = userID, limit, offset
			return []domain.Task{{ID: 1, Title: "Buy milk"}}, nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/tasks?user_id=5&limit=10&offset=20", nil))
	rw := httptest.NewRecorder()

	h.GetTasks(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if receivedUserID == nil || *receivedUserID != 5 {
		t.Fatalf("expected userID=5, got %v", receivedUserID)
	}
	if receivedLimit == nil || *receivedLimit != 10 {
		t.Fatalf("expected limit=10, got %v", receivedLimit)
	}
	if receivedOffset == nil || *receivedOffset != 20 {
		t.Fatalf("expected offset=20, got %v", receivedOffset)
	}

	var resp GetTasksResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 task in response, got %d", len(resp))
	}
}

func TestTasksHandler_GetTasks_InvalidQueryParam(t *testing.T) {
	called := false
	svc := &mockTasksService{
		getTasksFn: func(_ context.Context, _ *int, _ *int, _ *int) ([]domain.Task, error) {
			called = true
			return nil, nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/tasks?limit=abc", nil))
	rw := httptest.NewRecorder()

	h.GetTasks(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rw.Code)
	}
	if called {
		t.Fatalf("expected service NOT to be called")
	}
}


func TestTasksHandler_PatchTask_Valid(t *testing.T) {
	var receivedID int
	var receivedPatch domain.TaskPatch
	svc := &mockTasksService{
		patchTaskFn: func(_ context.Context, id int, patch domain.TaskPatch) (domain.Task, error) {
			receivedID = id
			receivedPatch = patch
			return domain.Task{ID: id, Title: "New title"}, nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := newJSONRequest(http.MethodPatch, "/tasks/3", `{"title":"New title"}`)
	req.SetPathValue("id", "3")
	rw := httptest.NewRecorder()

	h.PatchTask(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if receivedID != 3 {
		t.Fatalf("expected id=3, got %d", receivedID)
	}
	if !receivedPatch.Title.Set || *receivedPatch.Title.Value != "New title" {
		t.Fatalf("expected patch title 'New title', got %+v", receivedPatch.Title)
	}
}

func TestTasksHandler_PatchTask_InvalidBody(t *testing.T) {
	called := false
	svc := &mockTasksService{
		patchTaskFn: func(_ context.Context, id int, patch domain.TaskPatch) (domain.Task, error) {
			called = true
			return domain.Task{}, nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := newJSONRequest(http.MethodPatch, "/tasks/3", `{"title":null}`)
	req.SetPathValue("id", "3")
	rw := httptest.NewRecorder()

	h.PatchTask(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rw.Code)
	}
	if called {
		t.Fatalf("expected service NOT to be called")
	}
}

func TestTasksHandler_PatchTask_Conflict(t *testing.T) {
	svc := &mockTasksService{
		patchTaskFn: func(_ context.Context, id int, patch domain.TaskPatch) (domain.Task, error) {
			return domain.Task{}, core_errors.ErrConflict
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := newJSONRequest(http.MethodPatch, "/tasks/3", `{"title":"New title"}`)
	req.SetPathValue("id", "3")
	rw := httptest.NewRecorder()

	h.PatchTask(rw, req)

	if rw.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rw.Code)
	}
}


func TestTasksHandler_DeleteTask_Valid(t *testing.T) {
	var receivedID int
	svc := &mockTasksService{
		deleteTaskFn: func(_ context.Context, id int) error {
			receivedID = id
			return nil
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodDelete, "/tasks/9", nil))
	req.SetPathValue("id", "9")
	rw := httptest.NewRecorder()

	h.DeleteTask(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rw.Code)
	}
	if receivedID != 9 {
		t.Fatalf("expected id=9, got %d", receivedID)
	}
}

func TestTasksHandler_DeleteTask_NotFound(t *testing.T) {
	svc := &mockTasksService{
		deleteTaskFn: func(_ context.Context, id int) error {
			return core_errors.ErrNotFound
		},
	}
	h := NewTasksHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodDelete, "/tasks/999", nil))
	req.SetPathValue("id", "999")
	rw := httptest.NewRecorder()

	h.DeleteTask(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rw.Code)
	}

	var body core_http_response.ErrorResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid JSON error body: %v", err)
	}
	if body.Message == "" {
		t.Fatalf("expected non-empty message")
	}
}
