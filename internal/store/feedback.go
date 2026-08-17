package store

import (
	"fmt"
	"sort"

	"volunteertraining/internal/domain"
)

func (s *Store) SaveFeedback(feedback domain.Feedback) error {
	if err := feedback.Validate(); err != nil {
		return err
	}
	feedback.ID = stableID("feedback", feedback.VolunteerID, feedback.VideoID, feedback.CreatedAt.UTC().String())
	if err := s.put(bucketFeedback, feedback.ID, feedback); err != nil {
		return fmt.Errorf("save feedback: %w", err)
	}
	return nil
}

func (s *Store) ListFeedback(videoID string) ([]domain.Feedback, error) {
	items, err := s.list(bucketFeedback, func() any { return &domain.Feedback{} }, func(value any) string { return value.(*domain.Feedback).ID })
	if err != nil {
		return nil, fmt.Errorf("list feedback: %w", err)
	}
	feedback := make([]domain.Feedback, 0)
	for _, item := range items {
		value := *item.(*domain.Feedback)
		if videoID == "" || value.VideoID == videoID {
			feedback = append(feedback, value)
		}
	}
	sort.Slice(feedback, func(i, j int) bool { return feedback[i].CreatedAt.Before(feedback[j].CreatedAt) })
	return feedback, nil
}

func (s *Store) AverageRating(videoID string) (float64, error) {
	items, err := s.ListFeedback(videoID)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, nil
	}
	total := 0
	for _, item := range items {
		total += item.Rating
	}
	return float64(total) / float64(len(items)), nil
}
