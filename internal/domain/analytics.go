package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type VideoFilter struct {
	Query     string
	Role      Role
	Published *bool
	Required  *bool
}

func (f VideoFilter) Match(video TrainingVideo) bool {
	if f.Role.Valid() && video.Role != f.Role {
		return false
	}
	if f.Published != nil && video.Published != *f.Published {
		return false
	}
	if f.Required != nil && video.Required != *f.Required {
		return false
	}
	query := strings.ToLower(strings.TrimSpace(f.Query))
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(video.ID), query) || strings.Contains(strings.ToLower(video.Title), query) || strings.Contains(strings.ToLower(video.Description), query) || strings.Contains(strings.ToLower(video.ExamTip), query)
}

type TrainingPlan struct {
	Volunteer User
	Videos    []TrainingVideo
	Progress  map[string]ViewingProgress
	CreatedAt time.Time
}

func (p TrainingPlan) CompletedCount() int {
	count := 0
	for _, video := range p.Videos {
		if progress, ok := p.Progress[video.ID]; ok && progress.Completed {
			count++
		}
	}
	return count
}

func (p TrainingPlan) Pending() []TrainingVideo {
	pending := make([]TrainingVideo, 0)
	for _, video := range p.Videos {
		progress, ok := p.Progress[video.ID]
		if !ok || !progress.Completed {
			pending = append(pending, video)
		}
	}
	return pending
}

func (p TrainingPlan) Complete() bool {
	return len(p.Videos) > 0 && p.CompletedCount() == len(p.Videos)
}

func BuildTrainingPlan(user User, videos []TrainingVideo, progress []ViewingProgress, createdAt time.Time) TrainingPlan {
	progressMap := make(map[string]ViewingProgress, len(progress))
	for _, item := range progress {
		progressMap[item.VideoID] = item
	}
	selected := make([]TrainingVideo, 0, len(videos))
	for _, video := range videos {
		if RequiredForRole(video, user.Role) {
			selected = append(selected, video)
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].ID < selected[j].ID })
	return TrainingPlan{Volunteer: user, Videos: selected, Progress: progressMap, CreatedAt: createdAt}
}

type VolunteerProgressView struct {
	VolunteerID string
	Volunteer   string
	Role        Role
	VideoID     string
	VideoTitle  string
	Completed   bool
	WatchedSec  int
	UpdatedAt   time.Time
}

func (v VolunteerProgressView) Status() string {
	if v.Completed {
		return "completed"
	}
	return "pending"
}

type AuditQuery struct {
	ActorID    string
	Action     string
	EntityType string
	Since      time.Time
	Until      time.Time
}

func (q AuditQuery) Match(record AuditRecord) bool {
	if q.ActorID != "" && record.ActorID != q.ActorID {
		return false
	}
	if q.Action != "" && record.Action != q.Action {
		return false
	}
	if q.EntityType != "" && record.EntityType != q.EntityType {
		return false
	}
	if !q.Since.IsZero() && record.CreatedAt.Before(q.Since) {
		return false
	}
	if !q.Until.IsZero() && record.CreatedAt.After(q.Until) {
		return false
	}
	return true
}

func ValidateVideoFilter(filter VideoFilter) error {
	if filter.Role != "" && !filter.Role.Valid() {
		return fmt.Errorf("%w: invalid video filter role", ErrInvalid)
	}
	return nil
}

func SortVideos(videos []TrainingVideo) []TrainingVideo {
	result := append([]TrainingVideo(nil), videos...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Role == result[j].Role {
			return result[i].Title < result[j].Title
		}
		return result[i].Role < result[j].Role
	})
	return result
}

func SummarizeCompletion(role Role, total, completed int) CompletionSummary {
	if completed > total {
		completed = total
	}
	if completed < 0 {
		completed = 0
	}
	return CompletionSummary{Role: role, Total: total, Completed: completed, Outstanding: total - completed, Percent: CompletionPercent(total, completed)}
}

func DurationMinutes(video TrainingVideo) int {
	if video.DurationSec <= 0 {
		return 0
	}
	return (video.DurationSec + 59) / 60
}

func IsFreshProgress(progress ViewingProgress, at time.Time) bool {
	if progress.UpdatedAt.IsZero() || at.IsZero() {
		return false
	}
	return !progress.UpdatedAt.Before(at.Add(-24 * time.Hour))
}
