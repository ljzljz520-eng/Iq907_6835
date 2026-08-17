package store

import (
	"path/filepath"
	"testing"
	"time"

	"volunteertraining/internal/domain"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	database, err := Open(filepath.Join(t.TempDir(), "training.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func TestStorePersistsEntities(t *testing.T) {
	database := openTestStore(t)
	when := time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)
	user := domain.User{ID: "u-1", Name: "One", Email: "one@example.test", Password: "pass", Role: domain.RoleGuide, Active: true, CreatedAt: when}
	video := domain.TrainingVideo{ID: "v-1", Title: "Brief", Role: domain.RoleGuide, DurationSec: 60, URL: "https://test/video", Required: true, Published: true, CreatedAt: when}
	if err := database.SaveUser(user); err != nil {
		t.Fatal(err)
	}
	if err := database.SaveVideo(video); err != nil {
		t.Fatal(err)
	}
	progress := domain.ViewingProgress{VolunteerID: user.ID, VideoID: video.ID, WatchedSec: 60, Completed: true, CompletedAt: when, UpdatedAt: when}
	audit := domain.AuditRecord{ActorID: user.ID, Action: "training.completed", EntityType: "ViewingProgress", EntityID: video.ID, CreatedAt: when}
	if err := database.RecordCompletion(progress, audit); err != nil {
		t.Fatal(err)
	}
	if err := database.SaveFeedback(domain.Feedback{VolunteerID: user.ID, VideoID: video.ID, Rating: 4, Comment: "clear", CreatedAt: when}); err != nil {
		t.Fatal(err)
	}
	if _, err := database.FindProgress(user.ID, video.ID); err != nil {
		t.Fatal(err)
	}
	if count, err := database.AuditCount("ViewingProgress", video.ID); err != nil || count != 1 {
		t.Fatalf("audit count=%d err=%v", count, err)
	}
}

func TestRoleSummaryCountsActiveVolunteers(t *testing.T) {
	database := openTestStore(t)
	when := time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)
	for _, user := range []domain.User{{ID: "u-a", Name: "A", Email: "a@example.test", Password: "pass", Role: domain.RoleGreeter, Active: true, CreatedAt: when}, {ID: "u-b", Name: "B", Email: "b@example.test", Password: "pass", Role: domain.RoleGreeter, Active: false, CreatedAt: when}} {
		if err := database.SaveUser(user); err != nil {
			t.Fatal(err)
		}
	}
	if err := database.SaveVideo(domain.TrainingVideo{ID: "v-a", Title: "A", Role: domain.RoleGreeter, DurationSec: 20, Required: true, Published: true, CreatedAt: when}); err != nil {
		t.Fatal(err)
	}
	summary, err := database.RoleSummary(domain.RoleGreeter)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total != 1 || summary.Outstanding != 1 {
		t.Fatalf("summary=%+v", summary)
	}
}
