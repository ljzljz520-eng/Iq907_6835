package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound      = errors.New("entity not found")
	ErrInvalid       = errors.New("invalid entity")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrAlreadyExists = errors.New("entity already exists")
	ErrConflict      = errors.New("state conflict")
	ErrCancelled     = errors.New("operation cancelled")
)

type Role string

const (
	RoleGreeter     Role = "greeter"
	RoleGuide       Role = "guide"
	RoleCoordinator Role = "coordinator"
	RoleTrainer     Role = "trainer"
)

func (r Role) Valid() bool {
	switch r {
	case RoleGreeter, RoleGuide, RoleCoordinator, RoleTrainer:
		return true
	default:
		return false
	}
}

type User struct {
	ID          string
	Name        string
	Email       string
	Password    string
	Role        Role
	Active      bool
	CreatedAt   time.Time
	LastLoginAt time.Time
}

func (u User) Validate() error {
	if strings.TrimSpace(u.ID) == "" || strings.TrimSpace(u.Name) == "" {
		return fmt.Errorf("%w: user id and name are required", ErrInvalid)
	}
	if !strings.Contains(u.Email, "@") {
		return fmt.Errorf("%w: email is invalid", ErrInvalid)
	}
	if len(u.Password) < 4 {
		return fmt.Errorf("%w: password is too short", ErrInvalid)
	}
	if !u.Role.Valid() {
		return fmt.Errorf("%w: unknown role", ErrInvalid)
	}
	return nil
}

type TrainingVideo struct {
	ID          string
	Title       string
	Description string
	URL         string
	Role        Role
	DurationSec int
	ExamTip     string
	Required    bool
	Published   bool
	ImportedBy  string
	PublishedAt time.Time
	CreatedAt   time.Time
}

func (v TrainingVideo) Validate() error {
	if strings.TrimSpace(v.ID) == "" || strings.TrimSpace(v.Title) == "" {
		return fmt.Errorf("%w: video id and title are required", ErrInvalid)
	}
	if !v.Role.Valid() || v.DurationSec <= 0 {
		return fmt.Errorf("%w: video role or duration is invalid", ErrInvalid)
	}
	return nil
}

type ViewingProgress struct {
	ID          string
	VolunteerID string
	VideoID     string
	WatchedSec  int
	Completed   bool
	CompletedAt time.Time
	UpdatedAt   time.Time
}

func (p ViewingProgress) Validate() error {
	if p.VolunteerID == "" || p.VideoID == "" || p.WatchedSec < 0 {
		return fmt.Errorf("%w: invalid progress", ErrInvalid)
	}
	return nil
}

type Feedback struct {
	ID          string
	VolunteerID string
	VideoID     string
	Rating      int
	Comment     string
	CreatedAt   time.Time
}

func (f Feedback) Validate() error {
	if f.VolunteerID == "" || f.VideoID == "" || f.Rating < 1 || f.Rating > 5 {
		return fmt.Errorf("%w: rating must be between one and five", ErrInvalid)
	}
	if len(strings.TrimSpace(f.Comment)) < 3 {
		return fmt.Errorf("%w: feedback comment is required", ErrInvalid)
	}
	return nil
}

type AuditRecord struct {
	ID         string
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	Details    string
	CreatedAt  time.Time
}

func (a AuditRecord) Validate() error {
	if a.ID == "" || a.ActorID == "" || a.Action == "" || a.EntityType == "" || a.EntityID == "" {
		return fmt.Errorf("%w: incomplete audit record", ErrInvalid)
	}
	return nil
}

type CompletionSummary struct {
	Role        Role
	Total       int
	Completed   int
	Outstanding int
	Percent     float64
}

type CSVImportResult struct {
	UsersImported  int
	VideosImported int
	Skipped        int
	Errors         []string
}

func (r CSVImportResult) Successful() bool { return len(r.Errors) == 0 }
