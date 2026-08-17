package catalog

import (
	"fmt"
	"sort"
	"strings"

	"volunteertraining/internal/domain"
)

type SearchResult struct {
	Video        domain.TrainingVideo
	DurationMin  int
	RoleLabel    string
	ExamTipMatch bool
}

func (s *Service) Search(filter domain.VideoFilter) ([]SearchResult, error) {
	items, err := s.store.SearchVideos(filter)
	if err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0, len(items))
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	for _, video := range items {
		results = append(results, SearchResult{Video: video, DurationMin: domain.DurationMinutes(video), RoleLabel: domain.RoleLabel(video.Role), ExamTipMatch: query != "" && strings.Contains(strings.ToLower(video.ExamTip), query)})
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].Video.Title < results[j].Video.Title })
	return results, nil
}

func (s *Service) UpdateExamTip(actor domain.User, videoID, tip string) error {
	if err := domain.Require(actor.Role, domain.PermissionPublish); err != nil {
		return err
	}
	if strings.TrimSpace(tip) == "" {
		return fmt.Errorf("%w: tip is empty", domain.ErrInvalid)
	}
	if err := s.store.UpdateVideoTip(videoID, tip); err != nil {
		return err
	}
	return s.store.SaveAudit(domain.AuditRecord{ActorID: actor.ID, Action: "catalog.tip.updated", EntityType: "TrainingVideo", EntityID: videoID, Details: tip, CreatedAt: s.now()})
}

func (s *Service) SetRequired(actor domain.User, videoID string, required bool) error {
	if err := domain.Require(actor.Role, domain.PermissionPublish); err != nil {
		return err
	}
	if err := s.store.SetVideoRequired(videoID, required); err != nil {
		return err
	}
	return s.store.SaveAudit(domain.AuditRecord{ActorID: actor.ID, Action: "catalog.requirement.changed", EntityType: "TrainingVideo", EntityID: videoID, Details: fmt.Sprintf("required=%t", required), CreatedAt: s.now()})
}
