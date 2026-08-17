package reporting

import (
	"encoding/csv"
	"fmt"
	"io"

	"volunteertraining/internal/domain"
)

func WriteCSV(writer io.Writer, summaries []domain.CompletionSummary) error {
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"role", "label", "total", "completed", "outstanding", "percent"}); err != nil {
		return err
	}
	for _, summary := range summaries {
		record := []string{string(summary.Role), domain.RoleLabel(summary.Role), fmt.Sprintf("%d", summary.Total), fmt.Sprintf("%d", summary.Completed), fmt.Sprintf("%d", summary.Outstanding), fmt.Sprintf("%.2f", summary.Percent)}
		if err := output.Write(record); err != nil {
			return err
		}
	}
	output.Flush()
	return output.Error()
}
