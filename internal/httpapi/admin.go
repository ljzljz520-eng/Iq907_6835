package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/reporting"
)

func (s *Server) searchCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	filter, err := videoFilterFromRequest(r)
	if err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	items, err := s.application.SearchCatalog(filter)
	if err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, items)
}

func videoFilterFromRequest(r *http.Request) (domain.VideoFilter, error) {
	filter := domain.VideoFilter{Query: r.URL.Query().Get("q")}
	role := strings.TrimSpace(r.URL.Query().Get("role"))
	if role != "" {
		filter.Role = domain.Role(role)
	}
	if value := strings.TrimSpace(r.URL.Query().Get("published")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return domain.VideoFilter{}, err
		}
		filter.Published = &parsed
	}
	if value := strings.TrimSpace(r.URL.Query().Get("required")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return domain.VideoFilter{}, err
		}
		filter.Required = &parsed
	}
	return filter, domain.ValidateVideoFilter(filter)
}

func (s *Server) updateTip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	var request struct {
		VideoID string `json:"video_id"`
		Tip     string `json:"tip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.application.UpdateExamTip(r.Header.Get("X-Actor-ID"), request.VideoID, request.Tip); err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) setRequired(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	var request struct {
		VideoID  string `json:"video_id"`
		Required bool   `json:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	actor, err := s.application.Auth.Find(r.Header.Get("X-Actor-ID"))
	if err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err := s.application.Catalog.SetRequired(actor, request.VideoID, request.Required); err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) volunteerReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	actor, err := s.application.Auth.Find(r.Header.Get("X-Actor-ID"))
	if err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err := domain.Require(actor.Role, domain.PermissionReport); err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	role := domain.Role(strings.TrimSpace(r.URL.Query().Get("role")))
	items, err := s.application.Reporting.Volunteers(role)
	if err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, items)
}

func (s *Server) exportReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	actor, err := s.application.Auth.Find(r.Header.Get("X-Actor-ID"))
	if err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err := domain.Require(actor.Role, domain.PermissionReport); err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	summaries, err := s.application.Reporting.AllRoles()
	if err != nil {
		s.respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=completion-report.csv")
	if err := reporting.WriteCSV(w, summaries); err != nil {
		return
	}
}

func (s *Server) auditTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	actor, err := s.application.Auth.Find(r.Header.Get("X-Actor-ID"))
	if err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	query, err := auditQueryFromRequest(r)
	if err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	items, err := s.application.Audit.Timeline(actor, query)
	if err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, items)
}

func auditQueryFromRequest(r *http.Request) (domain.AuditQuery, error) {
	query := domain.AuditQuery{ActorID: strings.TrimSpace(r.URL.Query().Get("actor_id")), Action: strings.TrimSpace(r.URL.Query().Get("action")), EntityType: strings.TrimSpace(r.URL.Query().Get("entity_type"))}
	if value := strings.TrimSpace(r.URL.Query().Get("since")); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return domain.AuditQuery{}, err
		}
		query.Since = parsed
	}
	if value := strings.TrimSpace(r.URL.Query().Get("until")); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return domain.AuditQuery{}, err
		}
		query.Until = parsed
	}
	return query, nil
}

func (s *Server) updateVolunteer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	var request struct {
		VolunteerID string `json:"volunteer_id"`
		Name        string `json:"name"`
		Active      *bool  `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	actorID := r.Header.Get("X-Actor-ID")
	if strings.TrimSpace(request.Name) != "" {
		if err := s.application.RenameVolunteer(actorID, request.VolunteerID, request.Name); err != nil {
			s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
	}
	if request.Active != nil && !*request.Active {
		if err := s.application.DeactivateVolunteer(actorID, request.VolunteerID); err != nil {
			s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
	}
	s.respond(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) snapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	actor, err := s.application.Auth.Find(r.Header.Get("X-Actor-ID"))
	if err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err := domain.Require(actor.Role, domain.PermissionReport); err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	data, err := s.application.Snapshot()
	if err != nil {
		s.respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=training-snapshot.json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
