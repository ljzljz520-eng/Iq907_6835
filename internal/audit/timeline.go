package audit

import (
	"sort"
	"strings"

	"volunteertraining/internal/domain"
)

func (s *Service) Timeline(actor domain.User, query domain.AuditQuery) ([]domain.AuditRecord, error) {
	if err := domain.Require(actor.Role, domain.PermissionAudit); err != nil {
		return nil, err
	}
	items, err := s.store.QueryAudit(query)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func GroupByAction(records []domain.AuditRecord) map[string]int {
	groups := make(map[string]int)
	for _, record := range records {
		key := strings.TrimSpace(record.Action)
		if key == "" {
			key = "unknown"
		}
		groups[key]++
	}
	return groups
}

func Actions(records []domain.AuditRecord) []string {
	groups := GroupByAction(records)
	actions := make([]string, 0, len(groups))
	for action := range groups {
		actions = append(actions, action)
	}
	sort.Strings(actions)
	return actions
}
