package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"volunteertraining/internal/app"
	"volunteertraining/internal/domain"
)

func TestHealthAndRequiredEndpoints(t *testing.T) {
	application, err := app.Open(filepath.Join(t.TempDir(), "http.db"), func() time.Time { return time.Unix(1, 0).UTC() })
	if err != nil {
		t.Fatal(err)
	}
	defer application.Close()
	if err := application.SeedFixtures(); err != nil {
		t.Fatal(err)
	}
	server := New(application)
	health := httptest.NewRecorder()
	server.Handler().ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status=%d", health.Code)
	}
	required := httptest.NewRecorder()
	server.Handler().ServeHTTP(required, httptest.NewRequest(http.MethodGet, "/volunteers/required?volunteer_id=u-alex", nil))
	if required.Code != http.StatusOK || !strings.Contains(required.Body.String(), "Warm Welcome") {
		t.Fatalf("required status=%d body=%s", required.Code, required.Body.String())
	}
	_ = domain.RoleGreeter
}
