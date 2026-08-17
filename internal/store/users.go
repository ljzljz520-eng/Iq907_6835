package store

import (
	"fmt"
	"sort"
	"time"

	"volunteertraining/internal/domain"
)

func (s *Store) SaveUser(user domain.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	if err := s.put(bucketUsers, user.ID, user); err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

func (s *Store) FindUser(id string) (domain.User, error) {
	var user domain.User
	if err := s.get(bucketUsers, id, &user); err != nil {
		return domain.User{}, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}

func (s *Store) FindUserByEmail(email string) (domain.User, error) {
	users, err := s.ListUsers()
	if err != nil {
		return domain.User{}, err
	}
	for _, user := range users {
		if domain.NormalizeEmail(user.Email) == domain.NormalizeEmail(email) {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

func (s *Store) ListUsers() ([]domain.User, error) {
	items, err := s.list(bucketUsers, func() any { return &domain.User{} }, func(value any) string { return value.(*domain.User).ID })
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]domain.User, 0, len(items))
	for _, item := range items {
		users = append(users, *item.(*domain.User))
	}
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	return users, nil
}

func (s *Store) SetLastLogin(id string, when time.Time) error {
	user, err := s.FindUser(id)
	if err != nil {
		return err
	}
	if when.IsZero() {
		return fmt.Errorf("%w: login time is empty", domain.ErrInvalid)
	}
	user.LastLoginAt = when
	return s.SaveUser(user)
}
