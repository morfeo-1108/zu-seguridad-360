package model

import "time"

type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

type InboxItem struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	CompanyID   string     `json:"company_id"`
	CompanyName string     `json:"company_name"`
	Type        string     `json:"type"` // incident | document | medical | drone | cert | copasst
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    Priority   `json:"priority"`
	ActionURL   string     `json:"action_url"`
	IsRead      bool       `json:"is_read"`
	IsDone      bool       `json:"is_done"`
	LegalImpact bool       `json:"legal_impact"`
	DeadlineAt  *time.Time `json:"deadline_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Score para priorización del inbox
func (i *InboxItem) PriorityScore() int {
	score := 0
	switch i.Priority {
	case PriorityCritical:
		score += 1000
	case PriorityHigh:
		score += 500
	case PriorityMedium:
		score += 200
	case PriorityLow:
		score += 50
	}
	if i.LegalImpact {
		score += 300
	}
	if i.DeadlineAt != nil {
		hoursLeft := time.Until(*i.DeadlineAt).Hours()
		if hoursLeft < 24 {
			score += 400
		} else if hoursLeft < 72 {
			score += 200
		}
	}
	return score
}
