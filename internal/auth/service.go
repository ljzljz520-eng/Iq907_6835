package auth

import (
	"fmt"
	"strings"
	"time"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/store"
)

type Service struct {
	users *store.Store
	now   func() time.Time
}

func NewService(users *store.Store, now func() time.Time) *Service {
	if now == nil {
		now = func() time.Time { return time.Unix(0, 0).UTC() }
	}
	return &Service{users: users, now: now}
}

func (s *Service) Register(user domain.User) error {
	user.ID = domain.NormalizeID(user.ID)
	user.Email = domain.NormalizeEmail(user.Email)
	if user.CreatedAt.IsZero() {
		user.CreatedAt = s.now()
	}
	if !user.Active {
		return fmt.Errorf("%w: inactive users cannot register", domain.ErrInvalid)
	}
	if _, err := s.users.FindUser(user.ID); err == nil {
		return domain.ErrAlreadyExists
	} else if err != nil && !isNotFound(err) {
		return err
	}
	return s.users.SaveUser(user)
}

func (s *Service) Login(email, password string) (domain.User, error) {
	if strings.TrimSpace(email) == "" || password == "" {
		return domain.User{}, fmt.Errorf("%w: credentials are required", domain.ErrUnauthorized)
	}
	user, err := s.users.FindUserByEmail(email)
	if err != nil {
		return domain.User{}, fmt.Errorf("%w: credentials are not valid", domain.ErrUnauthorized)
	}
	if !user.Active || user.Password != password {
		return domain.User{}, fmt.Errorf("%w: credentials are not valid", domain.ErrUnauthorized)
	}
	user.LastLoginAt = s.now()
	if err := s.users.SaveUser(user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *Service) Find(id string) (domain.User, error) {
	return s.users.FindUser(domain.NormalizeID(id))
}

func isNotFound(err error) bool { return strings.Contains(err.Error(), domain.ErrNotFound.Error()) }
