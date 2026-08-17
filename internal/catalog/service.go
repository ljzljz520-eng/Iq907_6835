package catalog

import (
	"fmt"
	"time"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/store"
)

type Service struct {
	store *store.Store
	now   func() time.Time
}

func NewService(database *store.Store, now func() time.Time) *Service {
	if now == nil {
		now = func() time.Time { return time.Unix(0, 0).UTC() }
	}
	return &Service{store: database, now: now}
}

func (s *Service) AddVideo(actor domain.User, video domain.TrainingVideo) error {
	if err := domain.Require(actor.Role, domain.PermissionImport); err != nil {
		return err
	}
	video.ImportedBy = actor.ID
	if video.CreatedAt.IsZero() {
		video.CreatedAt = s.now()
	}
	if video.Published && video.PublishedAt.IsZero() {
		video.PublishedAt = s.now()
	}
	if err := s.store.SaveVideo(video); err != nil {
		return err
	}
	return s.store.SaveAudit(domain.AuditRecord{ActorID: actor.ID, Action: "catalog.video.imported", EntityType: "TrainingVideo", EntityID: video.ID, Details: video.Title, CreatedAt: s.now()})
}

func (s *Service) Publish(actor domain.User, videoID string) error {
	if err := domain.Require(actor.Role, domain.PermissionPublish); err != nil {
		return err
	}
	if err := s.store.PublishVideo(videoID, actor.ID); err != nil {
		return err
	}
	return s.store.SaveAudit(domain.AuditRecord{ActorID: actor.ID, Action: "catalog.video.published", EntityType: "TrainingVideo", EntityID: videoID, CreatedAt: s.now()})
}

func (s *Service) Required(role domain.Role) ([]domain.TrainingVideo, error) {
	if !role.Valid() {
		return nil, fmt.Errorf("%w: invalid role", domain.ErrInvalid)
	}
	return s.store.ListRequiredVideos(role)
}

func (s *Service) All() ([]domain.TrainingVideo, error) { return s.store.ListVideos() }
