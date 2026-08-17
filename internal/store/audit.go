package store

import (
	"fmt"
	"sort"

	"volunteertraining/internal/domain"
)

func (s *Store) SaveAudit(record domain.AuditRecord) error {
	if record.ID == "" {
		record.ID = stableID("audit", record.ActorID, record.Action, record.EntityType, record.EntityID, record.CreatedAt.UTC().String())
	}
	if err := record.Validate(); err != nil {
		return err
	}
	if err := s.put(bucketAudit, record.ID, record); err != nil {
		return fmt.Errorf("save audit: %w", err)
	}
	return nil
}

func (s *Store) ListAudit(entityType, entityID string) ([]domain.AuditRecord, error) {
	items, err := s.list(bucketAudit, func() any { return &domain.AuditRecord{} }, func(value any) string { return value.(*domain.AuditRecord).ID })
	if err != nil {
		return nil, fmt.Errorf("list audit: %w", err)
	}
	records := make([]domain.AuditRecord, 0)
	for _, item := range items {
		value := *item.(*domain.AuditRecord)
		if (entityType == "" || value.EntityType == entityType) && (entityID == "" || value.EntityID == entityID) {
			records = append(records, value)
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].CreatedAt.Before(records[j].CreatedAt) })
	return records, nil
}

func (s *Store) AuditCount(entityType, entityID string) (int, error) {
	items, err := s.ListAudit(entityType, entityID)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}
