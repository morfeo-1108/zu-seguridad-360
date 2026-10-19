package worker

import (
	"context"
	"github.com/zu-seguridad/autopilot/config"
	"log"
)

type RiskMatrixWorker struct{ cfg *config.Config }

func NewRiskMatrixWorker(cfg *config.Config) *RiskMatrixWorker { return &RiskMatrixWorker{cfg: cfg} }
func (w *RiskMatrixWorker) Name() string                       { return "risk_matrix" }
func (w *RiskMatrixWorker) Schedule() string                   { return "0 0 2 * * *" }
func (w *RiskMatrixWorker) Run(ctx context.Context, companyID string) error {
	log.Printf("[risk_matrix] updating NR after incidents...")
	return nil
}
