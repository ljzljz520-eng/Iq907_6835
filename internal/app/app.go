package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"volunteertraining/internal/audit"
	"volunteertraining/internal/auth"
	"volunteertraining/internal/catalog"
	"volunteertraining/internal/domain"
	"volunteertraining/internal/reporting"
	"volunteertraining/internal/store"
	"volunteertraining/internal/training"
)

type Application struct {
	Store     *store.Store
	Auth      *auth.Service
	Catalog   *catalog.Service
	Training  *training.Service
	Reporting *reporting.Service
	Audit     *audit.Service
	now       func() time.Time
}

func Open(path string, now func() time.Time) (*Application, error) {
	database, err := store.Open(path)
	if err != nil {
		return nil, err
	}
	if now == nil {
		now = func() time.Time { return time.Unix(0, 0).UTC() }
	}
	return &Application{Store: database, Auth: auth.NewService(database, now), Catalog: catalog.NewService(database, now), Training: training.NewService(database, now), Reporting: reporting.NewService(database), Audit: audit.NewService(database, now), now: now}, nil
}

func (a *Application) Close() error {
	if a == nil || a.Store == nil {
		return nil
	}
	return a.Store.Close()
}

func (a *Application) SeedFixtures() error {
	for _, user := range domain.FixtureUsers() {
		if err := a.registerFixture(user); err != nil {
			return err
		}
	}
	trainer, err := a.Auth.Find("u-drew")
	if err != nil {
		return err
	}
	for _, video := range domain.FixtureVideos() {
		if err := a.Catalog.AddVideo(trainer, video); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) registerFixture(user domain.User) error {
	if _, err := a.Auth.Find(user.ID); err == nil {
		return nil
	}
	return a.Auth.Register(user)
}

func (a *Application) ImportCatalog(actorID string, reader io.Reader) (domain.CSVImportResult, error) {
	actor, err := a.Auth.Find(actorID)
	if err != nil {
		return domain.CSVImportResult{}, err
	}
	return a.Catalog.ImportCSV(actor, reader)
}

func (a *Application) VolunteerRequired(volunteerID string) ([]domain.TrainingVideo, error) {
	volunteer, err := a.Auth.Find(volunteerID)
	if err != nil {
		return nil, err
	}
	return a.Catalog.Required(volunteer.Role)
}

func (a *Application) Complete(ctx context.Context, volunteerID, videoID string, watchedSec int) (domain.ViewingProgress, error) {
	volunteer, err := a.Auth.Find(volunteerID)
	if err != nil {
		return domain.ViewingProgress{}, err
	}
	return a.Training.Complete(ctx, volunteer, videoID, watchedSec)
}

func (a *Application) CompleteBatch(ctx context.Context, volunteerID string, videoIDs []string, watchedSeconds map[string]int, afterCommit func(int)) error {
	volunteer, err := a.Auth.Find(volunteerID)
	if err != nil {
		return err
	}
	return a.Training.CompleteBatch(ctx, volunteer, videoIDs, watchedSeconds, afterCommit)
}

func (a *Application) SubmitFeedback(volunteerID, videoID string, rating int, comment string) (domain.Feedback, error) {
	volunteer, err := a.Auth.Find(volunteerID)
	if err != nil {
		return domain.Feedback{}, err
	}
	if err := domain.Require(volunteer.Role, domain.PermissionFeedback); err != nil {
		return domain.Feedback{}, err
	}
	if _, err := a.Store.FindVideo(videoID); err != nil {
		return domain.Feedback{}, err
	}
	feedback := domain.Feedback{VolunteerID: volunteerID, VideoID: videoID, Rating: rating, Comment: comment, CreatedAt: a.now()}
	if err := a.Store.SaveFeedback(feedback); err != nil {
		return domain.Feedback{}, err
	}
	if err := a.Store.SaveAudit(domain.AuditRecord{ActorID: volunteerID, Action: "feedback.submitted", EntityType: "Feedback", EntityID: videoID, Details: comment, CreatedAt: a.now()}); err != nil {
		return domain.Feedback{}, err
	}
	return feedback, nil
}

func (a *Application) RoleReport(actorID string, role domain.Role) (domain.CompletionSummary, error) {
	actor, err := a.Auth.Find(actorID)
	if err != nil {
		return domain.CompletionSummary{}, err
	}
	if err := domain.Require(actor.Role, domain.PermissionReport); err != nil {
		return domain.CompletionSummary{}, err
	}
	return a.Reporting.Role(role)
}

func (a *Application) AuditFor(actorID, entityType, entityID string) ([]domain.AuditRecord, error) {
	actor, err := a.Auth.Find(actorID)
	if err != nil {
		return nil, err
	}
	return a.Audit.ForEntity(actor, entityType, entityID)
}

func (a *Application) Validate() error {
	if a == nil || a.Store == nil {
		return fmt.Errorf("%w: application is nil", domain.ErrInvalid)
	}
	return a.Store.Health()
}
