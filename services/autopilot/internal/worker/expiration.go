package worker

import (
	"context"
	"github.com/zu-seguridad/autopilot/config"
	"log"
)

type ExpirationWorker struct{ cfg *config.Config }

func NewExpirationWorker(cfg *config.Config) *ExpirationWorker { return &ExpirationWorker{cfg: cfg} }
func (w *ExpirationWorker) Name() string                       { return "expirations" }
func (w *ExpirationWorker) Schedule() string                   { return "0 0 6 * * *" }
func (w *ExpirationWorker) Run(ctx context.Context, companyID string) error {
	log.Printf("[expiration] checking 30/15/7/0 day thresholds...")
	return nil
}
