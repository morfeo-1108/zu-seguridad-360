package worker

import (
	"context"
	"github.com/zu-seguridad/autopilot/config"
	"log"
)

type MonthlyReportWorker struct{ cfg *config.Config }

func NewMonthlyReportWorker(cfg *config.Config) *MonthlyReportWorker {
	return &MonthlyReportWorker{cfg: cfg}
}
func (w *MonthlyReportWorker) Name() string     { return "monthly_report" }
func (w *MonthlyReportWorker) Schedule() string { return "0 0 7 5 * *" }
func (w *MonthlyReportWorker) Run(ctx context.Context, companyID string) error {
	log.Printf("[monthly_report] generating SST reports...")
	return nil
}
