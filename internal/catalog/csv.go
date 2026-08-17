package catalog

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"volunteertraining/internal/domain"
)

func (s *Service) ImportCSV(actor domain.User, reader io.Reader) (domain.CSVImportResult, error) {
	if err := domain.Require(actor.Role, domain.PermissionImport); err != nil {
		return domain.CSVImportResult{}, err
	}
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	result := domain.CSVImportResult{Errors: make([]string, 0)}
	line := 0
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: %v", line, err))
			continue
		}
		if line == 1 && strings.EqualFold(strings.TrimSpace(record[0]), "id") {
			continue
		}
		if len(record) < 8 {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: expected eight columns", line))
			continue
		}
		video, parseErr := parseVideoRecord(record)
		if parseErr != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: %v", line, parseErr))
			continue
		}
		if err := s.AddVideo(actor, video); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: %v", line, err))
			continue
		}
		result.VideosImported++
	}
	return result, nil
}

func parseVideoRecord(record []string) (domain.TrainingVideo, error) {
	duration, err := strconv.Atoi(strings.TrimSpace(record[3]))
	if err != nil {
		return domain.TrainingVideo{}, fmt.Errorf("duration is invalid: %w", err)
	}
	role := domain.Role(strings.TrimSpace(record[2]))
	if !role.Valid() {
		return domain.TrainingVideo{}, fmt.Errorf("role %q is invalid", role)
	}
	required, err := strconv.ParseBool(strings.TrimSpace(record[5]))
	if err != nil {
		return domain.TrainingVideo{}, fmt.Errorf("required flag is invalid: %w", err)
	}
	return domain.TrainingVideo{ID: strings.TrimSpace(record[0]), Title: strings.TrimSpace(record[1]), Role: role, DurationSec: duration, URL: strings.TrimSpace(record[4]), Required: required, Published: true, ExamTip: strings.TrimSpace(record[6]), Description: strings.TrimSpace(record[7])}, nil
}
