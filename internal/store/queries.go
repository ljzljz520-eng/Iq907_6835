package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"volunteertraining/internal/domain"
)

func (s *Store) SearchVideos(filter domain.VideoFilter) ([]domain.TrainingVideo, error) {
	if err := domain.ValidateVideoFilter(filter); err != nil {
		return nil, err
	}
	videos, err := s.ListVideos()
	if err != nil {
		return nil, err
	}
	matched := make([]domain.TrainingVideo, 0, len(videos))
	for _, video := range videos {
		if filter.Match(video) {
			matched = append(matched, video)
		}
	}
	return domain.SortVideos(matched), nil
}

func (s *Store) ListFeedbackByVolunteer(volunteerID string) ([]domain.Feedback, error) {
	items, err := s.ListFeedback("")
	if err != nil {
		return nil, err
	}
	result := make([]domain.Feedback, 0)
	for _, item := range items {
		if item.VolunteerID == volunteerID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *Store) ListAuditByActor(actorID string) ([]domain.AuditRecord, error) {
	items, err := s.ListAudit("", "")
	if err != nil {
		return nil, err
	}
	result := make([]domain.AuditRecord, 0)
	for _, item := range items {
		if item.ActorID == actorID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *Store) QueryAudit(query domain.AuditQuery) ([]domain.AuditRecord, error) {
	items, err := s.ListAudit("", "")
	if err != nil {
		return nil, err
	}
	result := make([]domain.AuditRecord, 0)
	for _, item := range items {
		if query.Match(item) {
			result = append(result, item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result, nil
}

func (s *Store) HasCompletion(volunteerID, videoID string) (bool, error) {
	progress, err := s.FindProgress(volunteerID, videoID)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrNotFound.Error()) {
			return false, nil
		}
		return false, err
	}
	return progress.Completed, nil
}

func (s *Store) ProgressMatrix(role domain.Role) ([]domain.VolunteerProgressView, error) {
	users, err := s.ListUsers()
	if err != nil {
		return nil, err
	}
	videos, err := s.ListRequiredVideos(role)
	if err != nil {
		return nil, err
	}
	rows := make([]domain.VolunteerProgressView, 0)
	for _, user := range users {
		if user.Role != role || !user.Active {
			continue
		}
		for _, video := range videos {
			progress, progressErr := s.FindProgress(user.ID, video.ID)
			if progressErr != nil && !strings.Contains(progressErr.Error(), domain.ErrNotFound.Error()) {
				return nil, progressErr
			}
			rows = append(rows, domain.VolunteerProgressView{VolunteerID: user.ID, Volunteer: user.Name, Role: role, VideoID: video.ID, VideoTitle: video.Title, Completed: progressErr == nil && progress.Completed, WatchedSec: progress.WatchedSec, UpdatedAt: progress.UpdatedAt})
		}
	}
	return rows, nil
}

func (s *Store) ExportSnapshot() ([]byte, error) {
	users, err := s.ListUsers()
	if err != nil {
		return nil, err
	}
	videos, err := s.ListVideos()
	if err != nil {
		return nil, err
	}
	progress, err := s.ListProgress("")
	if err != nil {
		return nil, err
	}
	feedback, err := s.ListFeedback("")
	if err != nil {
		return nil, err
	}
	audits, err := s.ListAudit("", "")
	if err != nil {
		return nil, err
	}
	snapshot := struct {
		Users    []domain.User
		Videos   []domain.TrainingVideo
		Progress []domain.ViewingProgress
		Feedback []domain.Feedback
		Audit    []domain.AuditRecord
	}{users, videos, progress, feedback, audits}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode snapshot: %w", err)
	}
	return data, nil
}

func (s *Store) EntityCounts() (map[string]int, error) {
	users, err := s.ListUsers()
	if err != nil {
		return nil, err
	}
	videos, err := s.ListVideos()
	if err != nil {
		return nil, err
	}
	progress, err := s.ListProgress("")
	if err != nil {
		return nil, err
	}
	feedback, err := s.ListFeedback("")
	if err != nil {
		return nil, err
	}
	auditRecords, err := s.ListAudit("", "")
	if err != nil {
		return nil, err
	}
	return map[string]int{"User": len(users), "TrainingVideo": len(videos), "ViewingProgress": len(progress), "Feedback": len(feedback), "AuditRecord": len(auditRecords)}, nil
}
