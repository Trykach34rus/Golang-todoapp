package domain

import (
	"errors"
	"testing"

	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name:    "valid user without phone",
			user:    User{FullName: "Ivan Ivanov"},
			wantErr: false,
		},
		{
			name:    "valid user with phone",
			user:    User{FullName: "Ivan Ivanov", PhoneNumber: strPtr("+79876543210")},
			wantErr: false,
		},
		{
			name:    "full name too short",
			user:    User{FullName: "Iv"},
			wantErr: true,
		},
		{
			name:    "full name too long",
			user:    User{FullName: string(make([]rune, 101))},
			wantErr: true,
		},
		{
			name:    "phone missing plus prefix",
			user:    User{FullName: "Ivan Ivanov", PhoneNumber: strPtr("79876543210")},
			wantErr: true,
		},
		{
			name:    "phone with letters",
			user:    User{FullName: "Ivan Ivanov", PhoneNumber: strPtr("+7987abc3210")},
			wantErr: true,
		},
		{
			name:    "phone too short",
			user:    User{FullName: "Ivan Ivanov", PhoneNumber: strPtr("+7987")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tt.wantErr && err != nil && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected error to wrap ErrInvalidArgument, got: %v", err)
			}
		})
	}
}

func TestUser_ApplyPatch(t *testing.T) {
	t.Run("patch full name", func(t *testing.T) {
		user := User{FullName: "Old Name"}
		patch := NewUserPatch(
			Nullable[string]{Value: strPtr("New Name"), Set: true},
			Nullable[string]{},
		)

		if err := user.ApplyPatch(patch); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.FullName != "New Name" {
			t.Fatalf("expected 'New Name', got %q", user.FullName)
		}
	})

	t.Run("patch full name to nil is rejected", func(t *testing.T) {
		user := User{FullName: "Old Name"}
		patch := NewUserPatch(
			Nullable[string]{Value: nil, Set: true},
			Nullable[string]{},
		)

		if err := user.ApplyPatch(patch); err == nil {
			t.Fatalf("expected error when patching FullName to NULL")
		}
	})

	t.Run("patch phone number to nil clears it", func(t *testing.T) {
		user := User{FullName: "Old Name", PhoneNumber: strPtr("+79876543210")}
		patch := NewUserPatch(
			Nullable[string]{},
			Nullable[string]{Value: nil, Set: true},
		)

		if err := user.ApplyPatch(patch); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.PhoneNumber != nil {
			t.Fatalf("expected PhoneNumber to be nil, got %v", *user.PhoneNumber)
		}
	})

	t.Run("patch phone number to invalid format is rejected by post-patch validation", func(t *testing.T) {
		user := User{FullName: "Old Name"}
		patch := NewUserPatch(
			Nullable[string]{},
			Nullable[string]{Value: strPtr("not-a-phone"), Set: true},
		)

		if err := user.ApplyPatch(patch); err == nil {
			t.Fatalf("expected error for invalid phone format")
		}
	})
}
