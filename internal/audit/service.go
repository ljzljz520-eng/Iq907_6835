package audit

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

func (s *Service) Record(actor domain.User, action, entityType, entityID, details string) error {
	if err := domain.Require(actor.Role, domain.PermissionAudit); err != nil {
		return err
	}
	return s.store.SaveAudit(domain.AuditRecord{ActorID: actor.ID, Action: action, EntityType: entityType, EntityID: entityID, Details: details, CreatedAt: s.now()})
}

func (s *Service) ForEntity(actor domain.User, entityType, entityID string) ([]domain.AuditRecord, error) {
	if err := domain.Require(actor.Role, domain.PermissionAudit); err != nil {
		return nil, err
	}
	items, err := s.store.ListAudit(entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("query audit: %w", err)
	}
	return items, nil
}

func (s *Service) Count(actor domain.User, entityType, entityID string) (int, error) {
	items, err := s.ForEntity(actor, entityType, entityID)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}
