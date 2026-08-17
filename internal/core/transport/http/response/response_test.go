package core_http_response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	"go.uber.org/zap"
)

func newTestLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

// --- HTTPResponseHandler.JSONResponse ---------------------------------------

func TestJSONResponse_WritesStatusAndBody(t *testing.T) {
	rw := httptest.NewRecorder()
	handler := NewHTTPResponseHandler(newTestLogger(), rw)

	type payload struct {
		Name string `json:"name"`
	}

	handler.JSONResponse(payload{Name: "Ivan"}, http.StatusCreated)

	if rw.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rw.Code)
	}

	var got payload
	if err := json.Unmarshal(rw.Body.Bytes(), &got); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v, body: %s", err, rw.Body.String())
	}
	if got.Name != "Ivan" {
		t.Fatalf("expected name 'Ivan', got %q", got.Name)
	}
}

// --- HTTPResponseHandler.NoContentResponse ----------------------------------

func TestNoContentResponse(t *testing.T) {
	rw := httptest.NewRecorder()
	handler := NewHTTPResponseHandler(newTestLogger(), rw)

	handler.NoContentResponse()

	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rw.Code)
	}
	if rw.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rw.Body.String())
	}
}

// --- HTTPResponseHandler.HTMLResponse ----------------------------------------

func TestHTMLResponse(t *testing.T) {
	rw := httptest.NewRecorder()
	handler := NewHTTPResponseHandler(newTestLogger(), rw)

	handler.HTMLResponse([]byte("<h1>hello</h1>"))

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}
	if got := rw.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("expected text/html content type, got %q", got)
	}
	if rw.Body.String() != "<h1>hello</h1>" {
		t.Fatalf("expected body '<h1>hello</h1>', got %q", rw.Body.String())
	}
}

// --- HTTPResponseHandler.ErrorResponse ----------------------------------------

func TestErrorResponse_MapsErrorsToStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "invalid argument -> 400",
			err:        fmt.Errorf("bad input: %w", core_errors.ErrInvalidArgument),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found -> 404",
			err:        fmt.Errorf("task missing: %w", core_errors.ErrNotFound),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "conflict -> 409",
			err:        fmt.Errorf("version mismatch: %w", core_errors.ErrConflict),
			wantStatus: http.StatusConflict,
		},
		{
			name:       "unknown error -> 500",
			err:        fmt.Errorf("something exploded"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			handler := NewHTTPResponseHandler(newTestLogger(), rw)

			handler.ErrorResponse(tt.err, "operation failed")

			if rw.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rw.Code)
			}

			var body ErrorResponse
			if err := json.Unmarshal(rw.Body.Bytes(), &body); err != nil {
				t.Fatalf("expected valid JSON error body, got error: %v, body: %s", err, rw.Body.String())
			}
			if body.Message != "operation failed" {
				t.Fatalf("expected message 'operation failed', got %q", body.Message)
			}
			if body.Error == "" {
				t.Fatalf("expected non-empty error field")
			}
		})
	}
}

// --- HTTPResponseHandler.PanicResponse ----------------------------------------

func TestPanicResponse_Returns500WithMessage(t *testing.T) {
	rw := httptest.NewRecorder()
	handler := NewHTTPResponseHandler(newTestLogger(), rw)

	handler.PanicResponse("something went very wrong", "unexpected panic during request")

	if rw.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rw.Code)
	}

	var body ErrorResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid JSON error body, got error: %v, body: %s", err, rw.Body.String())
	}
	if body.Message != "unexpected panic during request" {
		t.Fatalf("expected message 'unexpected panic during request', got %q", body.Message)
	}
	if body.Error == "" {
		t.Fatalf("expected non-empty error field describing the panic")
	}
}

// --- ResponseWriter -----------------------------------------------------------

func TestResponseWriter_DefaultStatusIsOK(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	// WriteHeader was never called -- per the http.ResponseWriter contract
	// a response that only calls Write() implicitly gets a 200 status.
	if got := rw.GetStatusCode(); got != http.StatusOK {
		t.Fatalf("expected default status 200, got %d", got)
	}
}

func TestResponseWriter_CapturesWrittenStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.WriteHeader(http.StatusTeapot)

	if got := rw.GetStatusCode(); got != http.StatusTeapot {
		t.Fatalf("expected status 418, got %d", got)
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected underlying recorder to also see status 418, got %d", rec.Code)
	}
}
