package store

import (
	"fmt"
	"strings"

	"go.etcd.io/bbolt"

	"volunteertraining/internal/domain"
)

func (s *Store) UpdateUserName(id, name string) error {
	user, err := s.FindUser(id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is required", domain.ErrInvalid)
	}
	user.Name = strings.TrimSpace(name)
	return s.SaveUser(user)
}

func (s *Store) SetUserActive(id string, active bool) error {
	user, err := s.FindUser(id)
	if err != nil {
		return err
	}
	user.Active = active
	return s.SaveUser(user)
}

func (s *Store) UpdateVideoTip(id, tip string) error {
	video, err := s.FindVideo(id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(tip) == "" {
		return fmt.Errorf("%w: exam tip is required", domain.ErrInvalid)
	}
	video.ExamTip = strings.TrimSpace(tip)
	return s.SaveVideo(video)
}

func (s *Store) SetVideoRequired(id string, required bool) error {
	video, err := s.FindVideo(id)
	if err != nil {
		return err
	}
	video.Required = required
	return s.SaveVideo(video)
}

func (s *Store) RemoveVideo(id string) error {
	return s.update(func(tx *bbolt.Tx) error {
		if tx.Bucket(bucketVideos).Get([]byte(id)) == nil {
			return domain.ErrNotFound
		}
		return tx.Bucket(bucketVideos).Delete([]byte(id))
	})
}
