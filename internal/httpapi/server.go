package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"volunteertraining/internal/app"
	"volunteertraining/internal/domain"
)

type Server struct{ application *app.Application }

func New(application *app.Application) *Server { return &Server{application: application} }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/login", s.login)
	mux.HandleFunc("/catalog/import", s.importCatalog)
	mux.HandleFunc("/volunteers/required", s.required)
	mux.HandleFunc("/volunteers/complete", s.complete)
	mux.HandleFunc("/volunteers/feedback", s.feedback)
	mux.HandleFunc("/reports/role", s.report)
	mux.HandleFunc("/reports/volunteers", s.volunteerReport)
	mux.HandleFunc("/reports/export.csv", s.exportReport)
	mux.HandleFunc("/catalog/search", s.searchCatalog)
	mux.HandleFunc("/catalog/tip", s.updateTip)
	mux.HandleFunc("/catalog/required", s.setRequired)
	mux.HandleFunc("/audit/timeline", s.auditTimeline)
	mux.HandleFunc("/admin/volunteer", s.updateVolunteer)
	mux.HandleFunc("/admin/snapshot", s.snapshot)
	return logging(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	s.respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	user, err := s.application.Auth.Login(request.Email, request.Password)
	if err != nil {
		s.respond(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, user)
}

func (s *Server) importCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	actor := r.Header.Get("X-Actor-ID")
	result, err := s.application.ImportCatalog(actor, r.Body)
	if err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, result)
}

func (s *Server) required(w http.ResponseWriter, r *http.Request) {
	volunteerID := r.URL.Query().Get("volunteer_id")
	items, err := s.application.VolunteerRequired(volunteerID)
	if err != nil {
		s.respond(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, items)
}

func (s *Server) complete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	var request struct {
		VolunteerID string `json:"volunteer_id"`
		VideoID     string `json:"video_id"`
		WatchedSec  int    `json:"watched_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	progress, err := s.application.Complete(r.Context(), request.VolunteerID, request.VideoID, request.WatchedSec)
	if err != nil {
		s.respond(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, progress)
}

func (s *Server) feedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respond(w, http.StatusMethodNotAllowed, nil)
		return
	}
	var request struct {
		VolunteerID string `json:"volunteer_id"`
		VideoID     string `json:"video_id"`
		Rating      int    `json:"rating"`
		Comment     string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	feedback, err := s.application.SubmitFeedback(request.VolunteerID, request.VideoID, request.Rating, request.Comment)
	if err != nil {
		s.respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusCreated, feedback)
}

func (s *Server) report(w http.ResponseWriter, r *http.Request) {
	role := domain.Role(strings.TrimSpace(r.URL.Query().Get("role")))
	actor := r.Header.Get("X-Actor-ID")
	report, err := s.application.RoleReport(actor, role)
	if err != nil {
		s.respond(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	s.respond(w, http.StatusOK, report)
}

func (s *Server) respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(w).Encode(value)
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, 4<<20)
		}
		next.ServeHTTP(w, r)
	})
}

func ReadBody(r io.Reader) ([]byte, error) { return io.ReadAll(io.LimitReader(r, 4<<20)) }

func ParseInt(value string) (int, error) { return strconv.Atoi(strings.TrimSpace(value)) }

func Shutdown(ctx context.Context, server *http.Server) error { return server.Shutdown(ctx) }
