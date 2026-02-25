package entity

import (
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/valueobject"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name:    "valid user",
			user:    User{Email: "alice@example.com", FullName: "Alice", Role: valueobject.RoleAdmin},
			wantErr: false,
		},
		{
			name:    "missing email",
			user:    User{Email: "", FullName: "Alice", Role: valueobject.RoleAdmin},
			wantErr: true,
		},
		{
			name:    "missing full name",
			user:    User{Email: "alice@example.com", FullName: "", Role: valueobject.RoleAdmin},
			wantErr: true,
		},
		{
			name:    "invalid role",
			user:    User{Email: "alice@example.com", FullName: "Alice", Role: "superadmin"},
			wantErr: true,
		},
		{
			name:    "role manager valid",
			user:    User{Email: "bob@example.com", FullName: "Bob", Role: valueobject.RoleManager},
			wantErr: false,
		},
		{
			name:    "role viewer valid",
			user:    User{Email: "carol@example.com", FullName: "Carol", Role: valueobject.RoleViewer},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_CanLogin(t *testing.T) {
	tests := []struct {
		name string
		user User
		want bool
	}{
		{
			name: "active with hash",
			user: User{IsActive: true, PasswordHash: "$2a$10$hash"},
			want: true,
		},
		{
			name: "inactive with hash",
			user: User{IsActive: false, PasswordHash: "$2a$10$hash"},
			want: false,
		},
		{
			name: "active without hash",
			user: User{IsActive: true, PasswordHash: ""},
			want: false,
		},
		{
			name: "inactive without hash",
			user: User{IsActive: false, PasswordHash: ""},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.CanLogin(); got != tt.want {
				t.Fatalf("CanLogin() = %v, want %v", got, tt.want)
			}
		})
	}
}
