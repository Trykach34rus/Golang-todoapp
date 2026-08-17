package core_http_middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_http_response "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/response"
	"go.uber.org/zap"
)

func TestCORS_AllowedOrigin(t *testing.T) {
	nextCalled := false
	handler := CORS([]string{"http://example.com"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Origin", "http://example.com")
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if !nextCalled {
		t.Fatalf("expected next handler to be called")
	}
	if got := rw.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin=http://example.com, got %q", got)
	}
	if got := rw.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("expected Vary=Origin, got %q", got)
	}
	if got := rw.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatalf("expected Access-Control-Allow-Methods to be set")
	}
	if got := rw.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatalf("expected Access-Control-Allow-Headers to be set")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	nextCalled := false
	handler := CORS([]string{"http://example.com"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Origin", "http://evil.com")
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	// Request still reaches the handler -- CORS headers are what protect the
	// browser's ability to read the response, this middleware does not block
	// same-server requests from disallowed origins.
	if !nextCalled {
		t.Fatalf("expected next handler to still be called")
	}
	if got := rw.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no Access-Control-Allow-Origin header, got %q", got)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	handler := CORS([]string{"http://example.com"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil) // no Origin header at all
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if got := rw.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no Access-Control-Allow-Origin header, got %q", got)
	}
}

func TestCORS_PreflightOptions_AllowedOrigin(t *testing.T) {
	nextCalled := false
	handler := CORS([]string{"http://example.com"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		}),
	)

	req := httptest.NewRequest(http.MethodOptions, "/tasks", nil)
	req.Header.Set("Origin", "http://example.com")
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if nextCalled {
		t.Fatalf("expected preflight OPTIONS request to short-circuit before reaching next handler")
	}
	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rw.Code)
	}
	if got := rw.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Fatalf("expected CORS headers to be set on preflight response, got %q", got)
	}
}

func TestCORS_PreflightOptions_DisallowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://example.com"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)

	req := httptest.NewRequest(http.MethodOptions, "/tasks", nil)
	req.Header.Set("Origin", "http://evil.com")
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 even for disallowed origin preflight, got %d", rw.Code)
	}
	if got := rw.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS headers for disallowed origin, got %q", got)
	}
}


func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var receivedHeader string

	handler := RequestID()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeader = r.Header.Get(requestIDHeader)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil) // no X-Request-ID set
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if receivedHeader == "" {
		t.Fatalf("expected a generated request id on the incoming request, got empty string")
	}
	if receivedHeader == requestIDHeader {
		t.Fatalf("request id header value must not equal the header name itself ('%s')", requestIDHeader)
	}

	respHeader := rw.Header().Get(requestIDHeader)
	if respHeader != receivedHeader {
		t.Fatalf("expected response header to match request header, got request=%q response=%q", receivedHeader, respHeader)
	}
}

func TestRequestID_ReusesExisting(t *testing.T) {
	const existingID = "11111111-2222-3333-4444-555555555555"
	var receivedHeader string

	handler := RequestID()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeader = r.Header.Get(requestIDHeader)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set(requestIDHeader, existingID)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if receivedHeader != existingID {
		t.Fatalf("expected existing request id %q to be preserved, got %q", existingID, receivedHeader)
	}
	if got := rw.Header().Get(requestIDHeader); got != existingID {
		t.Fatalf("expected response header to reuse existing id %q, got %q", existingID, got)
	}
}

// --- Panic ------------------------------------------------------------

func newTestLoggerContext() context.Context {
	log := &core_logger.Logger{Logger: zap.NewNop()}
	return core_logger.ToContext(context.Background(), log)
}

func TestPanic_RecoversAndReturns500(t *testing.T) {
	handler := Panic()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("boom")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil).WithContext(newTestLoggerContext())
	rw := httptest.NewRecorder()

	// The panic must not escape ServeHTTP.
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rw.Code)
	}

	var body core_http_response.ErrorResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid JSON error response, got error: %v, body: %s", err, rw.Body.String())
	}
	if body.Error == "" {
		t.Fatalf("expected non-empty error field in response body")
	}
}

func TestPanic_NoPanicPassesThrough(t *testing.T) {
	handler := Panic()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil).WithContext(newTestLoggerContext())
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200 for non-panicking handler, got %d", rw.Code)
	}
	if rw.Body.String() != "ok" {
		t.Fatalf("expected body 'ok', got %q", rw.Body.String())
	}
}

func TestPanic_RecoversFromErrorPanic(t *testing.T) {
	handler := Panic()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(context.DeadlineExceeded) // panic with an error value, not just a string
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil).WithContext(newTestLoggerContext())
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rw.Code)
	}
}
