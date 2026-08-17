package domain

import "time"

func FixtureUsers() []User {
	base := time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)
	return []User{
		{ID: "u-alex", Name: "Alex Chen", Email: "alex@example.test", Password: "pass-alex", Role: RoleGreeter, Active: true, CreatedAt: base},
		{ID: "u-bea", Name: "Bea Singh", Email: "bea@example.test", Password: "pass-bea", Role: RoleGuide, Active: true, CreatedAt: base},
		{ID: "u-cam", Name: "Cam Rivera", Email: "cam@example.test", Password: "pass-cam", Role: RoleCoordinator, Active: true, CreatedAt: base},
		{ID: "u-drew", Name: "Drew Park", Email: "drew@example.test", Password: "pass-drew", Role: RoleTrainer, Active: true, CreatedAt: base},
	}
}

func FixtureVideos() []TrainingVideo {
	base := time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)
	return []TrainingVideo{
		{ID: "vid-greet-1", Title: "Warm Welcome", Description: "Greeting protocol", URL: "https://videos.test/warm", Role: RoleGreeter, DurationSec: 180, ExamTip: "Use names and eye contact", Required: true, Published: true, CreatedAt: base, PublishedAt: base},
		{ID: "vid-greet-2", Title: "Accessibility Basics", Description: "Inclusive service", URL: "https://videos.test/access", Role: RoleGreeter, DurationSec: 240, ExamTip: "Ask before assisting", Required: true, Published: true, CreatedAt: base, PublishedAt: base},
		{ID: "vid-guide-1", Title: "Route Briefing", Description: "Tour routes", URL: "https://videos.test/routes", Role: RoleGuide, DurationSec: 300, ExamTip: "Point out exits", Required: true, Published: true, CreatedAt: base, PublishedAt: base},
		{ID: "vid-coord-1", Title: "Incident Handoff", Description: "Escalation", URL: "https://videos.test/handoff", Role: RoleCoordinator, DurationSec: 360, ExamTip: "Record facts first", Required: true, Published: true, CreatedAt: base, PublishedAt: base},
	}
}
