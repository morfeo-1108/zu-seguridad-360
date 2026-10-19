package worker

import (
	"context"
	"github.com/zu-seguridad/autopilot/config"
	"log"
)

type COPASSTWorker struct{ cfg *config.Config }

func NewCOPASSTWorker(cfg *config.Config) *COPASSTWorker { return &COPASSTWorker{cfg: cfg} }
func (w *COPASSTWorker) Name() string                    { return "copasst_minutes" }
func (w *COPASSTWorker) Schedule() string                { return "0 0 8 1 * *" }
func (w *COPASSTWorker) Run(ctx context.Context, companyID string) error {
	log.Printf("[copasst] generating monthly COPASST minutes...")
	return nil
}
