package store

import (
	"fmt"
	"sort"

	"volunteertraining/internal/domain"
)

func (s *Store) SaveVideo(video domain.TrainingVideo) error {
	if err := video.Validate(); err != nil {
		return err
	}
	if err := s.put(bucketVideos, video.ID, video); err != nil {
		return fmt.Errorf("save video: %w", err)
	}
	return nil
}

func (s *Store) FindVideo(id string) (domain.TrainingVideo, error) {
	var video domain.TrainingVideo
	if err := s.get(bucketVideos, id, &video); err != nil {
		return domain.TrainingVideo{}, fmt.Errorf("find video: %w", err)
	}
	return video, nil
}

func (s *Store) ListVideos() ([]domain.TrainingVideo, error) {
	items, err := s.list(bucketVideos, func() any { return &domain.TrainingVideo{} }, func(value any) string { return value.(*domain.TrainingVideo).ID })
	if err != nil {
		return nil, fmt.Errorf("list videos: %w", err)
	}
	videos := make([]domain.TrainingVideo, 0, len(items))
	for _, item := range items {
		videos = append(videos, *item.(*domain.TrainingVideo))
	}
	sort.Slice(videos, func(i, j int) bool { return videos[i].ID < videos[j].ID })
	return videos, nil
}

func (s *Store) ListRequiredVideos(role domain.Role) ([]domain.TrainingVideo, error) {
	videos, err := s.ListVideos()
	if err != nil {
		return nil, err
	}
	selected := make([]domain.TrainingVideo, 0)
	for _, video := range videos {
		if domain.RequiredForRole(video, role) {
			selected = append(selected, video)
		}
	}
	return selected, nil
}

func (s *Store) PublishVideo(id, importer string) error {
	video, err := s.FindVideo(id)
	if err != nil {
		return err
	}
	if importer == "" {
		return fmt.Errorf("%w: importer is required", domain.ErrInvalid)
	}
	video.Published = true
	video.ImportedBy = importer
	return s.SaveVideo(video)
}
