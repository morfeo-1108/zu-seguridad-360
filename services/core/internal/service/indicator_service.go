package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/repository"
)

type IndicatorService struct {
	incidentRepo *repository.IncidentRepository
	employeeRepo *repository.EmployeeRepository
}

func NewIndicatorService(incidentRepo *repository.IncidentRepository, employeeRepo *repository.EmployeeRepository) *IndicatorService {
	return &IndicatorService{incidentRepo: incidentRepo, employeeRepo: employeeRepo}
}

// MonthlyIndicators computes IF (frequency) and IS (severity) for one
// calendar month using the standard Colombian SST constant (240,000 — man-
// hours equivalent to 100 full-time workers over a year). HHT (man-hours
// worked) is NOT derived from any timesheet or attendance record — this
// system has no attendance-tracking module — it is estimated as
// active_employees × standardMonthlyHours. That estimate is returned
// explicitly in the response so nobody mistakes it for a measured figure.
// Absenteeism is not computed here: it would require tracking paid leave,
// common-law incapacity, etc., none of which exists as a module yet: a
// partial number derived only from accident lost-days would misrepresent
// what absenteeism actually means.
func (s *IndicatorService) MonthlyIndicators(ctx context.Context, companyID, tenantID uuid.UUID, year, month, standardMonthlyHours int) (map[string]interface{}, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month debe estar entre 1 y 12")
	}
	if standardMonthlyHours <= 0 {
		standardMonthlyHours = 220 // aproximación común: ~8h/día x 27.5 días hábiles/mes
	}
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	activeEmployees, err := s.employeeRepo.CountByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	atCount, err := s.incidentRepo.CountByPeriod(ctx, companyID, from, to)
	if err != nil {
		return nil, err
	}
	lostDays, err := s.incidentRepo.SumLostDaysByPeriod(ctx, companyID, from, to)
	if err != nil {
		return nil, err
	}

	hht := float64(activeEmployees) * float64(standardMonthlyHours)
	var freqIndex, severityIndex float64
	if hht > 0 {
		freqIndex = float64(atCount) * 240000 / hht
		severityIndex = float64(lostDays) * 240000 / hht
	}

	return map[string]interface{}{
		"period":                 fmt.Sprintf("%04d-%02d", year, month),
		"active_employees":       activeEmployees,
		"standard_monthly_hours": standardMonthlyHours,
		"estimated_hht":          hht,
		"hht_is_estimated":       true,
		"work_accidents":         atCount,
		"lost_days":              lostDays,
		"frequency_index":        freqIndex,
		"severity_index":         severityIndex,
		"formula_note":           "IF = (accidentes × 240.000) / HHT; IS = (días perdidos × 240.000) / HHT. HHT estimado como trabajadores activos × horas mensuales estándar — no hay módulo de control de asistencia.",
		"absenteeism_note":       "No se calcula: requiere seguimiento de licencias e incapacidad común que no existe como módulo todavía.",
	}, nil
}
