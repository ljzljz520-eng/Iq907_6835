package training

import (
	"context"
	"fmt"

	"volunteertraining/internal/domain"
)

func (s *Service) CompleteBatch(ctx context.Context, volunteer domain.User, videoIDs []string, watchedSeconds map[string]int, afterCommit func(int)) error {
	if err := domain.Require(volunteer.Role, domain.PermissionComplete); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	for index, videoID := range videoIDs {
		// Honor cancellation at the boundary between records so the caller can
		// stop the batch before any further storage writes or audit records are
		// produced. Records already committed remain; those not yet started stay
		// untouched, and the cancellation cause is returned.
		if err := ctx.Err(); err != nil {
			return err
		}
		video, err := s.repository.FindVideo(videoID)
		if err != nil {
			return fmt.Errorf("batch video %s: %w", videoID, err)
		}
		watched := watchedSeconds[videoID]
		if watched < video.DurationSec {
			return fmt.Errorf("%w: video %s is incomplete", domain.ErrConflict, videoID)
		}
		completedAt := s.now()
		progress := domain.ViewingProgress{VolunteerID: volunteer.ID, VideoID: videoID, WatchedSec: watched, Completed: true, CompletedAt: completedAt, UpdatedAt: completedAt}
		audit := domain.AuditRecord{ActorID: volunteer.ID, Action: "training.completed.batch", EntityType: "ViewingProgress", EntityID: videoID, Details: video.Title, CreatedAt: completedAt}
		if err := s.repository.RecordCompletion(progress, audit); err != nil {
			return err
		}
		if afterCommit != nil {
			afterCommit(index)
		}
	}
	return ctx.Err()
}
