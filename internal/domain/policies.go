package domain

import (
	"fmt"
	"strings"
)

type Permission string

const (
	PermissionImport   Permission = "catalog.import"
	PermissionPublish  Permission = "catalog.publish"
	PermissionFeedback Permission = "feedback.submit"
	PermissionComplete Permission = "training.complete"
	PermissionReport   Permission = "report.read"
	PermissionAudit    Permission = "audit.read"
)

func Can(role Role, permission Permission) bool {
	switch permission {
	case PermissionImport, PermissionPublish, PermissionReport, PermissionAudit:
		return role == RoleTrainer || role == RoleCoordinator
	case PermissionFeedback, PermissionComplete:
		return role.Valid()
	default:
		return false
	}
}

func Require(role Role, permission Permission) error {
	if !Can(role, permission) {
		return fmt.Errorf("%w: %s cannot %s", ErrUnauthorized, role, permission)
	}
	return nil
}

func NormalizeID(value string) string { return strings.TrimSpace(strings.ToLower(value)) }

func NormalizeEmail(value string) string { return strings.TrimSpace(strings.ToLower(value)) }
