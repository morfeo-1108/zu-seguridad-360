package worker

import (
	"context"
	"github.com/zu-seguridad/autopilot/config"
	"log"
)

type EmployeeAlertWorker struct{ cfg *config.Config }

func NewEmployeeAlertWorker(cfg *config.Config) *EmployeeAlertWorker {
	return &EmployeeAlertWorker{cfg: cfg}
}
func (w *EmployeeAlertWorker) Name() string     { return "employee_alerts" }
func (w *EmployeeAlertWorker) Schedule() string { return "0 30 7 * * *" }
func (w *EmployeeAlertWorker) Run(ctx context.Context, companyID string) error {
	log.Printf("[employee_alerts] sending personalized expiration reminders...")
	return nil
}
