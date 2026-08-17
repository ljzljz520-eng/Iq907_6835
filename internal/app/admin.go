package app

import (
	"fmt"
	"strings"

	"volunteertraining/internal/catalog"
	"volunteertraining/internal/domain"
)

func (a *Application) RenameVolunteer(actorID, volunteerID, name string) error {
	actor, err := a.Auth.Find(actorID)
	if err != nil {
		return err
	}
	if err := domain.Require(actor.Role, domain.PermissionImport); err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is empty", domain.ErrInvalid)
	}
	if err := a.Store.UpdateUserName(volunteerID, name); err != nil {
		return err
	}
	return a.Store.SaveAudit(domain.AuditRecord{ActorID: actorID, Action: "user.renamed", EntityType: "User", EntityID: volunteerID, Details: name, CreatedAt: a.now()})
}

func (a *Application) DeactivateVolunteer(actorID, volunteerID string) error {
	actor, err := a.Auth.Find(actorID)
	if err != nil {
		return err
	}
	if err := domain.Require(actor.Role, domain.PermissionImport); err != nil {
		return err
	}
	if err := a.Store.SetUserActive(volunteerID, false); err != nil {
		return err
	}
	return a.Store.SaveAudit(domain.AuditRecord{ActorID: actorID, Action: "user.deactivated", EntityType: "User", EntityID: volunteerID, CreatedAt: a.now()})
}

func (a *Application) UpdateExamTip(actorID, videoID, tip string) error {
	actor, err := a.Auth.Find(actorID)
	if err != nil {
		return err
	}
	return a.Catalog.UpdateExamTip(actor, videoID, tip)
}

func (a *Application) SearchCatalog(filter domain.VideoFilter) ([]catalog.SearchResult, error) {
	return a.Catalog.Search(filter)
}

func (a *Application) Snapshot() ([]byte, error) { return a.Reporting.Snapshot() }
