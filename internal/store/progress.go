package store

import (
	"fmt"
	"sort"
	"time"

	"go.etcd.io/bbolt"

	"volunteertraining/internal/domain"
)

func (s *Store) RecordCompletion(progress domain.ViewingProgress, audit domain.AuditRecord) error {
	if err := progress.Validate(); err != nil {
		return err
	}
	if !progress.Completed {
		return fmt.Errorf("%w: completion must be true", domain.ErrInvalid)
	}
	if audit.Action == "" {
		return fmt.Errorf("%w: completion audit is required", domain.ErrInvalid)
	}
	progress.ID = stableID("progress", progress.VolunteerID, progress.VideoID)
	encodedProgress, err := marshal(progress)
	if err != nil {
		return err
	}
	audit.ID = stableID("audit", audit.ActorID, audit.Action, audit.EntityType, audit.EntityID, progress.CompletedAt.UTC().Format(time.RFC3339Nano))
	encodedAudit, err := marshal(audit)
	if err != nil {
		return err
	}
	return s.update(func(tx *bbolt.Tx) error {
		if err := tx.Bucket(bucketProgress).Put([]byte(progress.ID), encodedProgress); err != nil {
			return fmt.Errorf("write progress: %w", err)
		}
		if err := tx.Bucket(bucketAudit).Put([]byte(audit.ID), encodedAudit); err != nil {
			return fmt.Errorf("write completion audit: %w", err)
		}
		return nil
	})
}

func (s *Store) FindProgress(volunteerID, videoID string) (domain.ViewingProgress, error) {
	var progress domain.ViewingProgress
	if err := s.get(bucketProgress, stableID("progress", volunteerID, videoID), &progress); err != nil {
		return domain.ViewingProgress{}, fmt.Errorf("find progress: %w", err)
	}
	return progress, nil
}

func (s *Store) ListProgress(volunteerID string) ([]domain.ViewingProgress, error) {
	items, err := s.list(bucketProgress, func() any { return &domain.ViewingProgress{} }, func(value any) string { return value.(*domain.ViewingProgress).ID })
	if err != nil {
		return nil, fmt.Errorf("list progress: %w", err)
	}
	progress := make([]domain.ViewingProgress, 0)
	for _, item := range items {
		value := *item.(*domain.ViewingProgress)
		if volunteerID == "" || value.VolunteerID == volunteerID {
			progress = append(progress, value)
		}
	}
	sort.Slice(progress, func(i, j int) bool { return progress[i].VideoID < progress[j].VideoID })
	return progress, nil
}

func (s *Store) CountCompleted(volunteerID string) (int, error) {
	items, err := s.ListProgress(volunteerID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, item := range items {
		if item.Completed {
			count++
		}
	}
	return count, nil
}
