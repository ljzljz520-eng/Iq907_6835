package reporting

import (
	"fmt"
	"sort"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/store"
)

type Service struct{ store *store.Store }

func NewService(database *store.Store) *Service { return &Service{store: database} }

func (s *Service) Role(role domain.Role) (domain.CompletionSummary, error) {
	if !role.Valid() {
		return domain.CompletionSummary{}, fmt.Errorf("%w: invalid role", domain.ErrInvalid)
	}
	return s.store.RoleSummary(role)
}

func (s *Service) AllRoles() ([]domain.CompletionSummary, error) {
	roles := domain.AllRoles()
	result := make([]domain.CompletionSummary, 0, len(roles))
	for _, role := range roles {
		summary, err := s.Role(role)
		if err != nil {
			return nil, err
		}
		result = append(result, summary)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Role < result[j].Role })
	return result, nil
}

func (s *Service) CompletionForVolunteer(volunteerID string) (float64, error) {
	user, err := s.store.FindUser(volunteerID)
	if err != nil {
		return 0, err
	}
	videos, err := s.store.ListRequiredVideos(user.Role)
	if err != nil {
		return 0, err
	}
	completed, err := s.store.CountCompleted(volunteerID)
	if err != nil {
		return 0, err
	}
	return domain.CompletionPercent(len(videos), completed), nil
}
