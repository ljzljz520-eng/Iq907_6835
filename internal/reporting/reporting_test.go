package reporting

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/store"
)

func TestRoleReportAndCSV(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "report.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	when := time.Unix(10, 0).UTC()
	user := domain.User{ID: "vol", Name: "Volunteer", Email: "vol@example.test", Password: "pass", Role: domain.RoleGreeter, Active: true, CreatedAt: when}
	video := domain.TrainingVideo{ID: "vid", Title: "Welcome", Role: domain.RoleGreeter, DurationSec: 10, Required: true, Published: true, CreatedAt: when}
	if err := database.SaveUser(user); err != nil {
		t.Fatal(err)
	}
	if err := database.SaveVideo(video); err != nil {
		t.Fatal(err)
	}
	service := NewService(database)
	summary, err := service.Role(domain.RoleGreeter)
	if err != nil || summary.Total != 1 {
		t.Fatalf("summary=%+v err=%v", summary, err)
	}
	var output bytes.Buffer
	if err := WriteCSV(&output, []domain.CompletionSummary{summary}); err != nil {
		t.Fatal(err)
	}
	if output.Len() < 20 {
		t.Fatal("empty report")
	}
}
