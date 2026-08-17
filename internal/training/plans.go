package training

import (
	"context"
	"fmt"
	"sort"
	"time"

	"volunteertraining/internal/domain"
)

type PlanRepository interface {
	Repository
	ListVideos() ([]domain.TrainingVideo, error)
}

func BuildPlan(repository PlanRepository, volunteer domain.User, now time.Time) (domain.TrainingPlan, error) {
	videos, err := repository.ListVideos()
	if err != nil {
		return domain.TrainingPlan{}, err
	}
	progress, err := repository.ListProgress(volunteer.ID)
	if err != nil {
		return domain.TrainingPlan{}, err
	}
	return domain.BuildTrainingPlan(volunteer, videos, progress, now), nil
}

func NextRequired(plan domain.TrainingPlan) (domain.TrainingVideo, bool) {
	pending := plan.Pending()
	if len(pending) == 0 {
		return domain.TrainingVideo{}, false
	}
	sort.SliceStable(pending, func(i, j int) bool { return pending[i].ID < pending[j].ID })
	return pending[0], true
}

func ValidateBatchInputs(plan domain.TrainingPlan, ids []string, watched map[string]int) error {
	if len(ids) == 0 {
		return fmt.Errorf("%w: batch is empty", domain.ErrInvalid)
	}
	known := make(map[string]domain.TrainingVideo, len(plan.Videos))
	for _, video := range plan.Videos {
		known[video.ID] = video
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return fmt.Errorf("%w: duplicate video %s", domain.ErrConflict, id)
		}
		seen[id] = true
		video, ok := known[id]
		if !ok {
			return fmt.Errorf("%w: video %s is not required", domain.ErrConflict, id)
		}
		if watched[id] < video.DurationSec {
			return fmt.Errorf("%w: video %s is incomplete", domain.ErrConflict, id)
		}
	}
	return nil
}

type BatchSummary struct {
	Requested int
	Recorded  int
	Cancelled bool
	Elapsed   time.Duration
}

func RunValidatedBatch(ctx context.Context, service *Service, plan domain.TrainingPlan, ids []string, watched map[string]int, after func(int)) (BatchSummary, error) {
	started := plan.CreatedAt
	if err := ValidateBatchInputs(plan, ids, watched); err != nil {
		return BatchSummary{}, err
	}
	err := service.CompleteBatch(ctx, plan.Volunteer, ids, watched, after)
	finished := service.now()
	summary := BatchSummary{Requested: len(ids), Elapsed: finished.Sub(started)}
	if err != nil {
		summary.Cancelled = err == context.Canceled || err == context.DeadlineExceeded
	}
	if !summary.Cancelled {
		summary.Recorded = len(ids)
	}
	return summary, err
}
