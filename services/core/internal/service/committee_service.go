package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type CommitteeService struct{ db *gorm.DB }

func NewCommitteeService(db *gorm.DB) *CommitteeService { return &CommitteeService{db: db} }

type CommitteeMemberInput struct {
	EmployeeID  string `json:"employee_id" validate:"required"`
	Role        string `json:"role" validate:"required"` // "representante_empleador" | "representante_trabajadores"
	IsAlternate bool   `json:"is_alternate"`
}

type CreateCommitteeInput struct {
	Type        string                 `json:"type" validate:"required,oneof=copasst vigia"`
	PeriodStart string                 `json:"period_start" validate:"required"`
	PeriodEnd   string                 `json:"period_end" validate:"required"`
	Members     []CommitteeMemberInput `json:"members" validate:"required,min=2"`
}

// CreateCommittee validates every member's employee_id belongs to the
// caller's active company before persisting the roster.
func (s *CommitteeService) CreateCommittee(ctx context.Context, input CreateCommitteeInput, companyID, tenantID uuid.UUID) (*model.Committee, error) {
	periodStart, err := time.Parse("2006-01-02", input.PeriodStart)
	if err != nil {
		return nil, fmt.Errorf("period_start debe ser YYYY-MM-DD")
	}
	periodEnd, err := time.Parse("2006-01-02", input.PeriodEnd)
	if err != nil || !periodEnd.After(periodStart) {
		return nil, fmt.Errorf("period_end debe ser posterior a period_start")
	}

	members := model.JSONBArray{}
	employeeIDs := make([]uuid.UUID, 0, len(input.Members))
	for _, m := range input.Members {
		id, err := uuid.Parse(m.EmployeeID)
		if err != nil {
			return nil, fmt.Errorf("employee_id inválido en la lista de miembros")
		}
		employeeIDs = append(employeeIDs, id)
		members = append(members, map[string]interface{}{"employee_id": m.EmployeeID, "role": m.Role, "is_alternate": m.IsAlternate})
	}

	committee := &model.Committee{
		TenantID: tenantID, CompanyID: companyID, Type: model.CommitteeType(input.Type),
		PeriodStart: periodStart, PeriodEnd: periodEnd, Members: members,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		unique := map[uuid.UUID]bool{}
		for _, id := range employeeIDs {
			unique[id] = true
		}
		var count int64
		if err := tx.Table("employees").Where("id IN ? AND tenant_id = ? AND company_id = ?", employeeIDs, tenantID, companyID).
			Distinct("id").Count(&count).Error; err != nil {
			return err
		}
		if int(count) != len(unique) {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(committee).Error
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("uno o más miembros no pertenecen a la empresa activa")
		}
		return nil, err
	}
	return committee, nil
}

func (s *CommitteeService) List(ctx context.Context, companyID, tenantID uuid.UUID) ([]model.Committee, error) {
	var committees []model.Committee
	err := s.db.WithContext(ctx).Where("company_id = ? AND tenant_id = ?", companyID, tenantID).Order("period_start DESC").Find(&committees).Error
	return committees, err
}

type CreateMeetingInput struct {
	HeldAt    string                   `json:"held_at" validate:"required"`
	Agenda    string                   `json:"agenda" validate:"required"`
	Minutes   string                   `json:"minutes" validate:"required"`
	Attendees []string                 `json:"attendees"`
	Decisions []map[string]interface{} `json:"decisions"`
}

func (s *CommitteeService) CreateMeeting(ctx context.Context, committeeID uuid.UUID, input CreateMeetingInput, companyID, tenantID, recordedBy uuid.UUID) (*model.CommitteeMeeting, error) {
	heldAt, err := time.Parse("2006-01-02", input.HeldAt)
	if err != nil {
		heldAt, err = time.Parse(time.RFC3339, input.HeldAt)
		if err != nil {
			return nil, fmt.Errorf("held_at debe ser YYYY-MM-DD o RFC3339")
		}
	}

	attendees := model.JSONBArray{}
	for _, a := range input.Attendees {
		attendees = append(attendees, a)
	}
	decisions := model.JSONBArray{}
	for _, d := range input.Decisions {
		decisions = append(decisions, d)
	}

	meeting := &model.CommitteeMeeting{
		TenantID: tenantID, CompanyID: companyID, CommitteeID: committeeID,
		HeldAt: heldAt, Agenda: input.Agenda, Minutes: input.Minutes,
		Attendees: attendees, Decisions: decisions, RecordedBy: recordedBy,
	}

	var count int64
	if err := s.db.WithContext(ctx).Table("committees").Where("id = ? AND tenant_id = ? AND company_id = ?", committeeID, tenantID, companyID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, fmt.Errorf("el comité no pertenece a la empresa activa")
	}
	return meeting, s.db.WithContext(ctx).Create(meeting).Error
}

func (s *CommitteeService) ListMeetings(ctx context.Context, committeeID, companyID, tenantID uuid.UUID) ([]model.CommitteeMeeting, error) {
	var meetings []model.CommitteeMeeting
	err := s.db.WithContext(ctx).
		Where("committee_id = ? AND company_id = ? AND tenant_id = ?", committeeID, companyID, tenantID).
		Order("held_at DESC").Find(&meetings).Error
	return meetings, err
}
