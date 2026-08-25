package statistics_tansport_http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	"go.uber.org/zap"
)

// --- mock service -----------------------------------------------------------

type mockStatisticsService struct {
	getStatisticsFn func(ctx context.Context, userID *int, from *time.Time, to *time.Time) (domain.Statistics, error)
}

func (m *mockStatisticsService) GetStatistics(ctx context.Context, userID *int, from *time.Time, to *time.Time) (domain.Statistics, error) {
	return m.getStatisticsFn(ctx, userID, from, to)
}

// --- helpers -----------------------------------------------------------

func withLoggerContext(req *http.Request) *http.Request {
	log := &core_logger.Logger{Logger: zap.NewNop()}
	return req.WithContext(core_logger.ToContext(req.Context(), log))
}

// --- GetStatistics -----------------------------------------------------------

func TestStatisticsHandler_GetStatistics_ParsesQueryParams(t *testing.T) {
	var receivedUserID *int
	var receivedFrom, receivedTo *time.Time

	rate := 50.0
	avg := 90 * time.Minute

	svc := &mockStatisticsService{
		getStatisticsFn: func(_ context.Context, userID *int, from *time.Time, to *time.Time) (domain.Statistics, error) {
			receivedUserID, receivedFrom, receivedTo = userID, from, to
			return domain.NewStatistics(10, 5, &rate, &avg), nil
		},
	}
	h := NewStatisticsService(svc)

	req := withLoggerContext(httptest.NewRequest(
		http.MethodGet,
		"/statistics?user_id=3&from=2026-01-01&to=2026-02-01",
		nil,
	))
	rw := httptest.NewRecorder()

	h.GetStatistics(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if receivedUserID == nil || *receivedUserID != 3 {
		t.Fatalf("expected userID=3, got %v", receivedUserID)
	}
	if receivedFrom == nil || receivedTo == nil {
		t.Fatalf("expected from/to to be parsed, got from=%v to=%v", receivedFrom, receivedTo)
	}

	var resp GetStatisticsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if resp.TasksCreated != 10 || resp.TasksCompleted != 5 {
		t.Fatalf("expected TasksCreated=10 TasksCompleted=5, got %+v", resp)
	}
	if resp.TasksCompletedRate == nil || *resp.TasksCompletedRate != 50.0 {
		t.Fatalf("expected TasksCompletedRate=50, got %v", resp.TasksCompletedRate)
	}
	if resp.TaskAverageCompletionTime == nil || *resp.TaskAverageCompletionTime != "1h30m0s" {
		t.Fatalf("expected avg completion time '1h30m0s', got %v", resp.TaskAverageCompletionTime)
	}
}

func TestStatisticsHandler_GetStatistics_NoParams(t *testing.T) {
	svc := &mockStatisticsService{
		getStatisticsFn: func(_ context.Context, userID *int, from *time.Time, to *time.Time) (domain.Statistics, error) {
			if userID != nil || from != nil || to != nil {
				t.Fatalf("expected all params nil, got userID=%v from=%v to=%v", userID, from, to)
			}
			return domain.NewStatistics(0, 0, nil, nil), nil
		},
	}
	h := NewStatisticsService(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/statistics", nil))
	rw := httptest.NewRecorder()

	h.GetStatistics(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}

	var resp GetStatisticsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if resp.TasksCompletedRate != nil {
		t.Fatalf("expected nil TasksCompletedRate, got %v", *resp.TasksCompletedRate)
	}
	if resp.TaskAverageCompletionTime != nil {
		t.Fatalf("expected nil TaskAverageCompletionTime, got %v", *resp.TaskAverageCompletionTime)
	}
}

func TestStatisticsHandler_GetStatistics_InvalidDateParam(t *testing.T) {
	called := false
	svc := &mockStatisticsService{
		getStatisticsFn: func(_ context.Context, _ *int, _ *time.Time, _ *time.Time) (domain.Statistics, error) {
			called = true
			return domain.Statistics{}, nil
		},
	}
	h := NewStatisticsService(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/statistics?from=not-a-date", nil))
	rw := httptest.NewRecorder()

	h.GetStatistics(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rw.Code)
	}
	if called {
		t.Fatalf("expected service NOT to be called")
	}
}

func TestStatisticsHandler_GetStatistics_ServiceError(t *testing.T) {
	svc := &mockStatisticsService{
		getStatisticsFn: func(_ context.Context, _ *int, _ *time.Time, _ *time.Time) (domain.Statistics, error) {
			return domain.Statistics{}, core_errors.ErrInvalidArgument
		},
	}
	h := NewStatisticsService(svc)

	req := withLoggerContext(httptest.NewRequest(http.MethodGet, "/statistics?from=2026-02-01&to=2026-01-01", nil))
	rw := httptest.NewRecorder()

	h.GetStatistics(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rw.Code)
	}
}
