package training

import (
	"context"
	"fmt"
	"time"

	"volunteertraining/internal/domain"
)

type Repository interface {
	FindVideo(string) (domain.TrainingVideo, error)
	RecordCompletion(domain.ViewingProgress, domain.AuditRecord) error
	FindProgress(string, string) (domain.ViewingProgress, error)
	ListProgress(string) ([]domain.ViewingProgress, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository, now func() time.Time) *Service {
	if now == nil {
		now = func() time.Time { return time.Unix(0, 0).UTC() }
	}
	return &Service{repository: repository, now: now}
}

func (s *Service) Complete(ctx context.Context, volunteer domain.User, videoID string, watchedSec int) (domain.ViewingProgress, error) {
	if err := domain.Require(volunteer.Role, domain.PermissionComplete); err != nil {
		return domain.ViewingProgress{}, err
	}
	video, err := s.repository.FindVideo(videoID)
	if err != nil {
		return domain.ViewingProgress{}, err
	}
	if watchedSec < video.DurationSec {
		return domain.ViewingProgress{}, fmt.Errorf("%w: watch %d of %d seconds", domain.ErrConflict, watchedSec, video.DurationSec)
	}
	if err := ctx.Err(); err != nil {
		return domain.ViewingProgress{}, err
	}
	completedAt := s.now()
	progress := domain.ViewingProgress{VolunteerID: volunteer.ID, VideoID: video.ID, WatchedSec: watchedSec, Completed: true, CompletedAt: completedAt, UpdatedAt: completedAt}
	audit := domain.AuditRecord{ActorID: volunteer.ID, Action: "training.completed", EntityType: "ViewingProgress", EntityID: video.ID, Details: video.Title, CreatedAt: completedAt}
	if err := s.repository.RecordCompletion(progress, audit); err != nil {
		return domain.ViewingProgress{}, err
	}
	return progress, nil
}

func (s *Service) Progress(volunteerID, videoID string) (domain.ViewingProgress, error) {
	return s.repository.FindProgress(volunteerID, videoID)
}

func (s *Service) ListProgress(volunteerID string) ([]domain.ViewingProgress, error) {
	return s.repository.ListProgress(volunteerID)
}
