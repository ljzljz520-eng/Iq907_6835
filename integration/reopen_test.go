package integration

import (
	"path/filepath"
	"testing"
	"time"

	"volunteertraining/internal/app"
	"volunteertraining/internal/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reopen.db")
	clock := func() time.Time { return time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC) }
	first, err := app.Open(path, clock)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.SeedFixtures(); err != nil {
		t.Fatal(err)
	}
	if _, err := first.Complete(nilContext{}, "u-alex", "vid-greet-1", 180); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := app.Open(path, clock)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	progress, err := second.Store.FindProgress("u-alex", "vid-greet-1")
	if err != nil || !progress.Completed {
		t.Fatalf("progress=%+v err=%v", progress, err)
	}
	if _, err := second.Store.FindUser("u-alex"); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Store.FindVideo("vid-greet-1"); err != nil {
		t.Fatal(err)
	}
	if count, err := second.Store.AuditCount("ViewingProgress", "vid-greet-1"); err != nil || count != 1 {
		t.Fatalf("audit=%d err=%v", count, err)
	}
	_ = domain.RoleGreeter
}

type nilContext struct{}

func (nilContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (nilContext) Done() <-chan struct{}       { return nil }
func (nilContext) Err() error                  { return nil }
func (nilContext) Value(any) any               { return nil }
