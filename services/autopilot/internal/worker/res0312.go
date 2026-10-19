package worker

import (
	"context"
	"github.com/zu-seguridad/autopilot/config"
	"log"
)

type Res0312Worker struct{ cfg *config.Config }

func NewRes0312Worker(cfg *config.Config) *Res0312Worker { return &Res0312Worker{cfg: cfg} }
func (w *Res0312Worker) Name() string                    { return "res0312_report" }
func (w *Res0312Worker) Schedule() string                { return "0 0 8 28 * *" }
func (w *Res0312Worker) Run(ctx context.Context, companyID string) error {
	log.Printf("[res0312] generating Res.0312 compliance report...")
	return nil
}
