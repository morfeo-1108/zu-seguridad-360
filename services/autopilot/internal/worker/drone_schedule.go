package worker

import (
	"context"
	"github.com/zu-seguridad/autopilot/config"
	"log"
)

type DroneScheduleWorker struct{ cfg *config.Config }

func NewDroneScheduleWorker(cfg *config.Config) *DroneScheduleWorker {
	return &DroneScheduleWorker{cfg: cfg}
}
func (w *DroneScheduleWorker) Name() string     { return "drone_schedule" }
func (w *DroneScheduleWorker) Schedule() string { return "0 0 */6 * * *" }
func (w *DroneScheduleWorker) Run(ctx context.Context, companyID string) error {
	log.Printf("[drone_schedule] dispatching drone inspection missions...")
	return nil
}
