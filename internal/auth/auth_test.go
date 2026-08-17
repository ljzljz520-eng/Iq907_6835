package auth

import (
	"path/filepath"
	"testing"
	"time"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/store"
)

func TestRegisterAndLogin(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	clock := func() time.Time { return time.Date(2026, 2, 2, 9, 0, 0, 0, time.UTC) }
	service := NewService(database, clock)
	if err := service.Register(domain.User{ID: " U-1 ", Name: "Ada", Email: "ADA@EXAMPLE.TEST", Password: "pass", Role: domain.RoleGuide, Active: true}); err != nil {
		t.Fatal(err)
	}
	user, err := service.Login("ada@example.test", "pass")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "u-1" || user.LastLoginAt.IsZero() {
		t.Fatalf("user=%+v", user)
	}
	if _, err := service.Login("ada@example.test", "wrong"); err == nil {
		t.Fatal("expected bad login")
	}
}
