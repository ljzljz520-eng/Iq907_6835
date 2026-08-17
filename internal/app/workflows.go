package app

import (
	"context"
	"fmt"
	"io"

	"volunteertraining/internal/domain"
)

type WorkflowResult struct {
	Imported  domain.CSVImportResult
	Completed int
	Feedback  domain.Feedback
	Report    domain.CompletionSummary
	Audit     []domain.AuditRecord
}

func (a *Application) RunCatalogWorkflow(actorID string, csvData io.Reader) (WorkflowResult, error) {
	imported, err := a.ImportCatalog(actorID, csvData)
	if err != nil {
		return WorkflowResult{}, err
	}
	if !imported.Successful() {
		return WorkflowResult{Imported: imported}, fmt.Errorf("catalog import had errors")
	}
	return WorkflowResult{Imported: imported}, nil
}

func (a *Application) RunVolunteerWorkflow(ctx context.Context, volunteerID, videoID string, watched int, rating int, comment string) (WorkflowResult, error) {
	if _, err := a.Complete(ctx, volunteerID, videoID, watched); err != nil {
		return WorkflowResult{}, err
	}
	feedback, err := a.SubmitFeedback(volunteerID, videoID, rating, comment)
	if err != nil {
		return WorkflowResult{}, err
	}
	auditRecords, err := a.AuditFor("u-drew", "ViewingProgress", videoID)
	if err != nil {
		return WorkflowResult{}, err
	}
	return WorkflowResult{Completed: 1, Feedback: feedback, Audit: auditRecords}, nil
}

func (a *Application) RunReportingWorkflow(actorID string, role domain.Role) (WorkflowResult, error) {
	report, err := a.RoleReport(actorID, role)
	if err != nil {
		return WorkflowResult{}, err
	}
	auditRecords, err := a.AuditFor(actorID, "ViewingProgress", "")
	if err != nil {
		return WorkflowResult{}, err
	}
	return WorkflowResult{Report: report, Audit: auditRecords}, nil
}
