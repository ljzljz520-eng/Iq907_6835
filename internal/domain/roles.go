package domain

import "sort"

var roleLabels = map[Role]string{
	RoleGreeter:     "Welcome Desk",
	RoleGuide:       "Visitor Guide",
	RoleCoordinator: "Floor Coordinator",
	RoleTrainer:     "Training Lead",
}

func RoleLabel(role Role) string {
	if label, ok := roleLabels[role]; ok {
		return label
	}
	return "Unknown Role"
}

func AllRoles() []Role {
	roles := make([]Role, 0, len(roleLabels))
	for role := range roleLabels {
		roles = append(roles, role)
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i] < roles[j] })
	return roles
}

func RequiredForRole(v TrainingVideo, role Role) bool {
	return v.Role == role && v.Required && v.Published
}

func CompletionPercent(total, completed int) float64 {
	if total <= 0 {
		return 0
	}
	if completed < 0 {
		completed = 0
	}
	if completed > total {
		completed = total
	}
	return float64(completed) * 100 / float64(total)
}
