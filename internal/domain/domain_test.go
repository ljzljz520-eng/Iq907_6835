package domain

import "testing"

func TestRolePermissionsAndProgressPercent(t *testing.T) {
	if !Can(RoleTrainer, PermissionImport) || Can(RoleGreeter, PermissionImport) {
		t.Fatal("role permission mismatch")
	}
	if got := CompletionPercent(4, 3); got != 75 {
		t.Fatalf("percent: %v", got)
	}
	if RoleLabel(RoleGuide) == "Unknown Role" || len(AllRoles()) != 4 {
		t.Fatal("role catalog incomplete")
	}
}

func TestEntityValidation(t *testing.T) {
	if err := (User{ID: "u", Name: "n", Email: "bad", Password: "x", Role: RoleGuide}).Validate(); err == nil {
		t.Fatal("expected invalid user")
	}
	if err := (Feedback{VolunteerID: "u", VideoID: "v", Rating: 5, Comment: "okay"}).Validate(); err != nil {
		t.Fatal(err)
	}
}
