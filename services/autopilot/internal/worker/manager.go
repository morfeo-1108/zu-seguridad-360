package worker

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/zu-seguridad/autopilot/config"
)

type Worker interface {
	Name() string
	Schedule() string
	Run(ctx context.Context, companyID string) error
}

type WorkerManager struct {
	workers map[string]Worker
	cron    *cron.Cron
	cfg     *config.Config
	mu      sync.RWMutex
}

func NewWorkerManager(cfg *config.Config) *WorkerManager {
	return &WorkerManager{
		workers: make(map[string]Worker),
		cron:    cron.New(cron.WithSeconds()),
		cfg:     cfg,
	}
}

func (wm *WorkerManager) RegisterWorker(name string, w Worker) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workers[name] = w
}

func (wm *WorkerManager) Start(ctx context.Context) {
	wm.mu.RLock()
	for _, w := range wm.workers {
		workerCopy := w
		wm.cron.AddFunc(workerCopy.Schedule(), func() {
			if err := workerCopy.Run(ctx, "all"); err != nil {
				log.Printf("Worker %s failed: %v", workerCopy.Name(), err)
			}
		})
	}
	wm.mu.RUnlock()
	wm.cron.Start()
	<-ctx.Done()
	wm.cron.Stop()
}

func (wm *WorkerManager) TriggerNow(ctx context.Context, name string) error {
	wm.mu.RLock()
	w, ok := wm.workers[name]
	wm.mu.RUnlock()
	if !ok {
		return fmt.Errorf("worker %s not found", name)
	}
	go w.Run(ctx, "all")
	return nil
}

func (wm *WorkerManager) Count() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.workers)
}

func (wm *WorkerManager) Status() map[string]interface{} {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	status := make(map[string]interface{})
	for name, w := range wm.workers {
		status[name] = map[string]string{"schedule": w.Schedule(), "status": "active"}
	}
	return status
}
