package training

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/store"
)

type cancelAfterFirstRepository struct {
	base   Repository
	cancel context.CancelFunc
	calls  int
}

func (r *cancelAfterFirstRepository) FindVideo(id string) (domain.TrainingVideo, error) {
	return r.base.FindVideo(id)
}
func (r *cancelAfterFirstRepository) FindProgress(volunteerID, videoID string) (domain.ViewingProgress, error) {
	return r.base.FindProgress(volunteerID, videoID)
}
func (r *cancelAfterFirstRepository) ListProgress(volunteerID string) ([]domain.ViewingProgress, error) {
	return r.base.ListProgress(volunteerID)
}
func (r *cancelAfterFirstRepository) RecordCompletion(progress domain.ViewingProgress, audit domain.AuditRecord) error {
	if err := r.base.RecordCompletion(progress, audit); err != nil {
		return err
	}
	r.calls++
	if r.calls == 1 {
		r.cancel()
	}
	return nil
}

func TestBatchStopsWhenContextCancelled(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "cancel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	when := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	user := domain.User{ID: "vol", Name: "Volunteer", Email: "vol@example.test", Password: "pass", Role: domain.RoleGreeter, Active: true, CreatedAt: when}
	if err := database.SaveUser(user); err != nil {
		t.Fatal(err)
	}
	ids := []string{"one", "two", "three"}
	watched := map[string]int{}
	for _, id := range ids {
		if err := database.SaveVideo(domain.TrainingVideo{ID: id, Title: id, Role: domain.RoleGreeter, DurationSec: 10, Required: true, Published: true, CreatedAt: when}); err != nil {
			t.Fatal(err)
		}
		watched[id] = 10
	}
	ctx, cancel := context.WithCancel(context.Background())
	repository := &cancelAfterFirstRepository{base: database, cancel: cancel}
	service := NewService(repository, func() time.Time { return when })
	err = service.CompleteBatch(ctx, user, ids, watched, nil)
	if err == nil || err != context.Canceled {
		t.Fatalf("error=%v", err)
	}
	progress, err := database.ListProgress(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(progress) != 1 {
		t.Fatalf("progress=%d", len(progress))
	}
}
