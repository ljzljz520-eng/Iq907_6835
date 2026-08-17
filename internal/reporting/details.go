package reporting

import (
	"sort"
	"strings"

	"volunteertraining/internal/domain"
)

type VolunteerReport struct {
	VolunteerID string
	Name        string
	Role        domain.Role
	Total       int
	Completed   int
	Pending     int
	Percent     float64
	Feedback    int
}

func (s *Service) Volunteers(role domain.Role) ([]VolunteerReport, error) {
	users, err := s.store.ListUsers()
	if err != nil {
		return nil, err
	}
	videos, err := s.store.ListRequiredVideos(role)
	if err != nil {
		return nil, err
	}
	result := make([]VolunteerReport, 0)
	for _, user := range users {
		if user.Role != role || !user.Active {
			continue
		}
		progress, err := s.store.ListProgress(user.ID)
		if err != nil {
			return nil, err
		}
		completed := 0
		for _, item := range progress {
			if item.Completed {
				for _, video := range videos {
					if item.VideoID == video.ID {
						completed++
					}
				}
			}
		}
		feedback, err := s.store.ListFeedbackByVolunteer(user.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, VolunteerReport{VolunteerID: user.ID, Name: user.Name, Role: role, Total: len(videos), Completed: completed, Pending: len(videos) - completed, Percent: domain.CompletionPercent(len(videos), completed), Feedback: len(feedback)})
	}
	sort.Slice(result, func(i, j int) bool { return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name) })
	return result, nil
}

func (s *Service) CompletionMatrix(role domain.Role) ([]domain.VolunteerProgressView, error) {
	return s.store.ProgressMatrix(role)
}

func (s *Service) Snapshot() ([]byte, error) { return s.store.ExportSnapshot() }

func (s *Service) LowestCompletionRole() (domain.CompletionSummary, error) {
	summaries, err := s.AllRoles()
	if err != nil {
		return domain.CompletionSummary{}, err
	}
	if len(summaries) == 0 {
		return domain.CompletionSummary{}, nil
	}
	lowest := summaries[0]
	for _, summary := range summaries[1:] {
		if summary.Percent < lowest.Percent || (summary.Percent == lowest.Percent && summary.Role < lowest.Role) {
			lowest = summary
		}
	}
	return lowest, nil
}
