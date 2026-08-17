package core_http_request

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
)

// --- DecodeAndValidateRequest ---------------------------------------------

// sampleCustomValidatable mimics the real DTOs (e.g. PatchTaskRequest) that
// implement their own Validate() error and therefore bypass the generic
// struct validator.
type sampleCustomValidatable struct {
	Title string `json:"title"`
}

func (r *sampleCustomValidatable) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	return nil
}

// sampleTagValidated relies on the go-playground validator tags, exercised
// when the destination type does NOT implement the validatable interface.
type sampleTagValidated struct {
	Name string `json:"name" validate:"required"`
}

func newJSONRequest(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/whatever", bytes.NewBufferString(body))
}

func TestDecodeAndValidateRequest_MalformedJSON(t *testing.T) {
	req := newJSONRequest(`{"title":`) // truncated JSON

	var dest sampleCustomValidatable
	err := DecodeAndValidateRequest(req, &dest)

	if err == nil {
		t.Fatalf("expected error for malformed JSON")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
	}
}

func TestDecodeAndValidateRequest_CustomValidatable_Valid(t *testing.T) {
	req := newJSONRequest(`{"title":"Buy milk"}`)

	var dest sampleCustomValidatable
	if err := DecodeAndValidateRequest(req, &dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest.Title != "Buy milk" {
		t.Fatalf("expected Title='Buy milk', got %q", dest.Title)
	}
}

func TestDecodeAndValidateRequest_CustomValidatable_InvalidatesViaOwnValidate(t *testing.T) {
	req := newJSONRequest(`{"title":""}`) // empty title -> custom Validate() fails

	var dest sampleCustomValidatable
	err := DecodeAndValidateRequest(req, &dest)

	if err == nil {
		t.Fatalf("expected error from custom Validate()")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
	}
}

func TestDecodeAndValidateRequest_StructTagValidation_Valid(t *testing.T) {
	req := newJSONRequest(`{"name":"Ivan"}`)

	var dest sampleTagValidated
	if err := DecodeAndValidateRequest(req, &dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeAndValidateRequest_StructTagValidation_MissingRequiredField(t *testing.T) {
	req := newJSONRequest(`{}`) // name is required but missing

	var dest sampleTagValidated
	err := DecodeAndValidateRequest(req, &dest)

	if err == nil {
		t.Fatalf("expected validation error for missing required field")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
	}
}

// --- GetIntPathValue --------------------------------------------------------

func TestGetIntPathValue_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks/", nil)
	// no path value set at all

	_, err := GetIntPathValue(req, "id")
	if err == nil {
		t.Fatalf("expected error for missing path value")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
	}
}

func TestGetIntPathValue_NotAnInteger(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
	req.SetPathValue("id", "abc")

	_, err := GetIntPathValue(req, "id")
	if err == nil {
		t.Fatalf("expected error for non-integer path value")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
	}
}

func TestGetIntPathValue_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks/42", nil)
	req.SetPathValue("id", "42")

	val, err := GetIntPathValue(req, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

// --- GetIntQueryParam --------------------------------------------------------

func TestGetIntQueryParam_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)

	val, err := GetIntQueryParam(req, "limit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil for missing query param, got %v", *val)
	}
}

func TestGetIntQueryParam_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks?limit=abc", nil)

	_, err := GetIntQueryParam(req, "limit")
	if err == nil {
		t.Fatalf("expected error for non-integer query param")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
	}
}

func TestGetIntQueryParam_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks?limit=10", nil)

	val, err := GetIntQueryParam(req, "limit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val == nil || *val != 10 {
		t.Fatalf("expected 10, got %v", val)
	}
}

// --- GetDateQueryParam --------------------------------------------------------

func TestGetDateQueryParam_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/statistics", nil)

	val, err := GetDateQueryParam(req, "from")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil for missing query param, got %v", *val)
	}
}

func TestGetDateQueryParam_InvalidFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/statistics?from=15-01-2026", nil) // wrong layout

	_, err := GetDateQueryParam(req, "from")
	if err == nil {
		t.Fatalf("expected error for invalid date format")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
	}
}

func TestGetDateQueryParam_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/statistics?from=2026-01-15", nil)

	val, err := GetDateQueryParam(req, "from")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val == nil {
		t.Fatalf("expected non-nil date")
	}

	want := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	if !val.Equal(want) {
		t.Fatalf("expected %v, got %v", want, *val)
	}
}
