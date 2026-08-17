package integration

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"volunteertraining/internal/app"
	"volunteertraining/internal/domain"
)

func newApplication(t *testing.T) *app.Application {
	t.Helper()
	application, err := app.Open(filepath.Join(t.TempDir(), "integration.db"), func() time.Time { return time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = application.Close() })
	return application
}

func TestWorkflowAdminImportsTrainingCatalog(t *testing.T) {
	application := newApplication(t)
	if err := application.SeedFixtures(); err != nil {
		t.Fatal(err)
	}
	data := "id,title,role,duration,url,required,tip,description\nvid-new,New Brief,guide,80,https://test/new,true,Remember exits,New route\n"
	result, err := application.RunCatalogWorkflow("u-drew", strings.NewReader(data))
	if err != nil || result.Imported.VideosImported != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	items, err := application.VolunteerRequired("u-bea")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("required=%+v", items)
	}
}

func TestWorkflowVolunteerCompletesAndSubmitsFeedback(t *testing.T) {
	application := newApplication(t)
	if err := application.SeedFixtures(); err != nil {
		t.Fatal(err)
	}
	result, err := application.RunVolunteerWorkflow(context.Background(), "u-alex", "vid-greet-1", 180, 5, "Clear and welcoming")
	if err != nil {
		t.Fatal(err)
	}
	if result.Completed != 1 || result.Feedback.Rating != 5 || len(result.Audit) != 1 {
		t.Fatalf("result=%+v", result)
	}
}

func TestWorkflowAdminReviewsRoleCompletion(t *testing.T) {
	application := newApplication(t)
	if err := application.SeedFixtures(); err != nil {
		t.Fatal(err)
	}
	if _, err := application.Complete(context.Background(), "u-alex", "vid-greet-1", 180); err != nil {
		t.Fatal(err)
	}
	result, err := application.RunReportingWorkflow("u-drew", domain.RoleGreeter)
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Total != 2 || result.Report.Completed != 1 || result.Report.Outstanding != 1 {
		t.Fatalf("report=%+v", result.Report)
	}
}
