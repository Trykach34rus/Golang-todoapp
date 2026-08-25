package users_transport_http

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
	"go.uber.org/zap"
)

// --- mock service -----------------------------------------------------------

type mockUsersService struct {
	createUserFn func(ctx context.Context, user domain.User) (domain.User, error)
	getUserFn    func(ctx context.Context, id int) (domain.User, error)
	getUsersFn   func(ctx context.Context, limit *int, offset *int) ([]domain.User, error)
	deleteUserFn func(ctx context.Context, id int) error
	patchUserFn  func(ctx context.Context, id int, patch domain.UserPatch) (domain.User, error)
}

func (m *mockUsersService) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return m.createUserFn(ctx, user)
}

func (m *mockUsersService) GetUsers(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
	return m.getUsersFn(ctx, limit, offset)
}

func (m *mockUsersService) GetUser(ctx context.Context, id int) (domain.User, error) {
	return m.getUserFn(ctx, id)
}

func (m *mockUsersService) DeleteUser(ctx context.Context, id int) error {
	return m.deleteUserFn(ctx, id)
}

func (m *mockUsersService) PatchUser(ctx context.Context, id int, patch domain.UserPatch) (domain.User, error) {
	return m.patchUserFn(ctx, id, patch)
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



func TestUsersHandler_CreateUser_Valid(t *testing.T) {
	var receivedUser domain.User
	svc := &mockUsersService{
		createUserFn: func(_ context.Context, user domain.User) (domain.User, error) {
			receivedUser = user
			user.ID = 1
			return user, nil
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := newJSONRequest(http.MethodPost, "/users", `{"full_name":"Ivan Ivanov"}`)
	rw := httptest.NewRecorder()

	h.CreateUser(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rw.Code, rw.Body.String())
	}
	if receivedUser.FullName != "Ivan Ivanov" {
		t.Fatalf("expected service to receive full name 'Ivan Ivanov', got %q", receivedUser.FullName)
	}

	var resp CreateUserResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("expected ID 1, got %d", resp.ID)
	}
}

func TestUsersHandler_CreateUser_InvalidBody(t *testing.T) {
	called := false
	svc := &mockUsersService{
		createUserFn: func(_ context.Context, user domain.User) (domain.User, error) {
			called = true
			return user, nil
		},
	}
	h := NewUsersHTTPHandler(svc)

	// full_name too short (< 3 chars per validate tag)
	req := newJSONRequest(http.MethodPost, "/users", `{"full_name":"Iv"}`)
	rw := httptest.NewRecorder()

	h.CreateUser(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rw.Code)
	}
	if called {
		t.Fatalf("expected service NOT to be called")
	}
}


func TestUsersHandler_GetUser_Valid(t *testing.T) {
	svc := &mockUsersService{
		getUserFn: func(_ context.Context, id int) (domain.User, error) {
			return domain.User{ID: id, FullName: "Ivan Ivanov"}, nil
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/users/4", nil))
	req.SetPathValue("id", "4")
	rw := httptest.NewRecorder()

	h.GetUser(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}

	var resp GetUserResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if resp.ID != 4 {
		t.Fatalf("expected ID 4, got %d", resp.ID)
	}
}

func TestUsersHandler_GetUser_NotFound(t *testing.T) {
	svc := &mockUsersService{
		getUserFn: func(_ context.Context, id int) (domain.User, error) {
			return domain.User{}, core_errors.ErrNotFound
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/users/999", nil))
	req.SetPathValue("id", "999")
	rw := httptest.NewRecorder()

	h.GetUser(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rw.Code)
	}
}


func TestUsersHandler_GetUsers_ParsesQueryParamsAndReturnsList(t *testing.T) {
	var receivedLimit, receivedOffset *int
	svc := &mockUsersService{
		getUsersFn: func(_ context.Context, limit *int, offset *int) ([]domain.User, error) {
			receivedLimit, receivedOffset = limit, offset
			return []domain.User{{ID: 1, FullName: "Ivan Ivanov"}}, nil
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/users?limit=10&offset=5", nil))
	rw := httptest.NewRecorder()

	h.GetUsers(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if receivedLimit == nil || *receivedLimit != 10 {
		t.Fatalf("expected limit=10, got %v", receivedLimit)
	}
	if receivedOffset == nil || *receivedOffset != 5 {
		t.Fatalf("expected offset=5, got %v", receivedOffset)
	}

	var resp GetUsersResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp))
	}
}


func TestUsersHandler_PatchUser_Valid(t *testing.T) {
	var receivedID int
	var receivedPatch domain.UserPatch
	svc := &mockUsersService{
		patchUserFn: func(_ context.Context, id int, patch domain.UserPatch) (domain.User, error) {
			receivedID = id
			receivedPatch = patch
			return domain.User{ID: id, FullName: "New Name"}, nil
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := newJSONRequest(http.MethodPatch, "/users/2", `{"full_name":"New Name"}`)
	req.SetPathValue("id", "2")
	rw := httptest.NewRecorder()

	h.PatchUser(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if receivedID != 2 {
		t.Fatalf("expected id=2, got %d", receivedID)
	}
	if !receivedPatch.FullName.Set || *receivedPatch.FullName.Value != "New Name" {
		t.Fatalf("expected patch full_name 'New Name', got %+v", receivedPatch.FullName)
	}
}

func TestUsersHandler_PatchUser_InvalidBody(t *testing.T) {
	called := false
	svc := &mockUsersService{
		patchUserFn: func(_ context.Context, id int, patch domain.UserPatch) (domain.User, error) {
			called = true
			return domain.User{}, nil
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := newJSONRequest(http.MethodPatch, "/users/2", `{"full_name":null}`)
	req.SetPathValue("id", "2")
	rw := httptest.NewRecorder()

	h.PatchUser(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rw.Code)
	}
	if called {
		t.Fatalf("expected service NOT to be called")
	}
}

func TestUsersHandler_PatchUser_Conflict(t *testing.T) {
	svc := &mockUsersService{
		patchUserFn: func(_ context.Context, id int, patch domain.UserPatch) (domain.User, error) {
			return domain.User{}, core_errors.ErrConflict
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := newJSONRequest(http.MethodPatch, "/users/2", `{"full_name":"New Name"}`)
	req.SetPathValue("id", "2")
	rw := httptest.NewRecorder()

	h.PatchUser(rw, req)

	if rw.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rw.Code)
	}
}


func TestUsersHandler_DeleteUser_Valid(t *testing.T) {
	var receivedID int
	svc := &mockUsersService{
		deleteUserFn: func(_ context.Context, id int) error {
			receivedID = id
			return nil
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodDelete, "/users/6", nil))
	req.SetPathValue("id", "6")
	rw := httptest.NewRecorder()

	h.DeleteUser(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rw.Code)
	}
	if receivedID != 6 {
		t.Fatalf("expected id=6, got %d", receivedID)
	}
}

func TestUsersHandler_DeleteUser_NotFound(t *testing.T) {
	svc := &mockUsersService{
		deleteUserFn: func(_ context.Context, id int) error {
			return core_errors.ErrNotFound
		},
	}
	h := NewUsersHTTPHandler(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodDelete, "/users/999", nil))
	req.SetPathValue("id", "999")
	rw := httptest.NewRecorder()

	h.DeleteUser(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rw.Code)
	}
}
