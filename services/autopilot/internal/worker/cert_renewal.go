package worker

import (
	"context"
	"github.com/zu-seguridad/autopilot/config"
	"log"
)

type CertRenewalWorker struct{ cfg *config.Config }

func NewCertRenewalWorker(cfg *config.Config) *CertRenewalWorker { return &CertRenewalWorker{cfg: cfg} }
func (w *CertRenewalWorker) Name() string                        { return "cert_renewal" }
func (w *CertRenewalWorker) Schedule() string                    { return "0 0 9 * * *" }
func (w *CertRenewalWorker) Run(ctx context.Context, companyID string) error {
	log.Printf("[cert_renewal] initiating certificate renewal flows...")
	return nil
}
