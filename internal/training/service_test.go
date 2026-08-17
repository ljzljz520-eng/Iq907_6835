package training

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/store"
)

func TestCompleteRequiresFullWatch(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "training.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	user := domain.User{ID: "vol", Name: "Volunteer", Email: "vol@example.test", Password: "pass", Role: domain.RoleGuide, Active: true, CreatedAt: time.Unix(1, 0).UTC()}
	video := domain.TrainingVideo{ID: "vid", Title: "Lesson", Role: domain.RoleGuide, DurationSec: 120, Required: true, Published: true, CreatedAt: time.Unix(1, 0).UTC()}
	if err := database.SaveUser(user); err != nil {
		t.Fatal(err)
	}
	if err := database.SaveVideo(video); err != nil {
		t.Fatal(err)
	}
	service := NewService(database, func() time.Time { return time.Unix(2, 0).UTC() })
	if _, err := service.Complete(context.Background(), user, video.ID, 119); err == nil {
		t.Fatal("expected incomplete error")
	}
	progress, err := service.Complete(context.Background(), user, video.ID, 120)
	if err != nil || !progress.Completed {
		t.Fatalf("progress=%+v err=%v", progress, err)
	}
}
