package store

import (
	"fmt"

	"go.etcd.io/bbolt"

	"volunteertraining/internal/domain"
)

func (s *Store) RoleSummary(role domain.Role) (domain.CompletionSummary, error) {
	users, err := s.ListUsers()
	if err != nil {
		return domain.CompletionSummary{}, err
	}
	videos, err := s.ListRequiredVideos(role)
	if err != nil {
		return domain.CompletionSummary{}, err
	}
	total := 0
	for _, user := range users {
		if user.Role != role || !user.Active {
			continue
		}
		for range videos {
			total++
		}
	}
	completed := 0
	for _, user := range users {
		if user.Role != role || !user.Active {
			continue
		}
		items, listErr := s.ListProgress(user.ID)
		if listErr != nil {
			return domain.CompletionSummary{}, listErr
		}
		for _, item := range items {
			if item.Completed {
				for _, video := range videos {
					if item.VideoID == video.ID {
						completed++
					}
				}
			}
		}
	}
	return domain.CompletionSummary{Role: role, Total: total, Completed: completed, Outstanding: total - completed, Percent: domain.CompletionPercent(total, completed)}, nil
}

func (s *Store) Health() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("%w: store is closed", domain.ErrConflict)
	}
	return s.view(func(*bbolt.Tx) error { return nil })
}
