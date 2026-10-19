package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

// Los 9 pasos del onboarding autónomo
// Desde RUT hasta primera ronda de dron
var OnboardingSteps = []Step{
	{ID: 1, Key: "rut_ocr", Title: "Procesamiento RUT", Description: "OCR del RUT para extraer NIT, razón social, CIIU y dirección", Auto: true},
	{ID: 2, Key: "risk_classification", Title: "Clasificación de Riesgo", Description: "Asignación automática de clase de riesgo ARL por código CIIU", Auto: true},
	{ID: 3, Key: "arl_config", Title: "Configuración ARL", Description: "Integración con ARL asignada y configuración de FURAT electrónico", Auto: false},
	{ID: 4, Key: "employee_import", Title: "Importación de Empleados", Description: "Carga del directorio de trabajadores desde CSV o Excel", Auto: false},
	{ID: 5, Key: "documents_gen", Title: "Generación Documental Inicial", Description: "Generación automática de los 6 documentos base del SG-SST", Auto: true},
	{ID: 6, Key: "risk_matrix_init", Title: "Matriz de Riesgos Inicial", Description: "Creación de la matriz GTC 45 con peligros estándar por sector CIIU", Auto: true},
	{ID: 7, Key: "training_schedule", Title: "Calendario de Capacitaciones", Description: "Programa anual de capacitaciones generado por IA según normas aplicables", Auto: true},
	{ID: 8, Key: "medical_schedule", Title: "Agendamiento Exámenes Médicos", Description: "Programación de exámenes de ingreso para todos los empleados", Auto: false},
	{ID: 9, Key: "drone_first_round", Title: "Primera Ronda de Dron", Description: "Misión de inspección inicial con reconocimiento de instalaciones", Auto: false},
}

type Step struct {
	ID          int    `json:"id"`
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Auto        bool   `json:"auto"`
}

type OnboardingState struct {
	CompanyID     string            `json:"company_id"`
	TenantID      string            `json:"tenant_id"`
	CurrentStep   int               `json:"current_step"`
	Completed     bool              `json:"completed"`
	StepStatus    map[string]string `json:"step_status"` // pending | running | done | failed
	StartedAt     time.Time         `json:"started_at"`
	CompletedAt   *time.Time        `json:"completed_at"`
	EstimatedDays int               `json:"estimated_days"`
}

type OnboardingService struct {
	db   *gorm.DB
	amqp interface {
		Publish(string, string, []byte) error
	}
}

func NewOnboardingService(db *gorm.DB, amqp interface {
	Publish(string, string, []byte) error
}) *OnboardingService {
	return &OnboardingService{db: db, amqp: amqp}
}

func (s *OnboardingService) Start(ctx context.Context, companyID, tenantID string) (*OnboardingState, error) {
	state := &OnboardingState{
		CompanyID:     companyID,
		TenantID:      tenantID,
		CurrentStep:   1,
		Completed:     false,
		StepStatus:    make(map[string]string),
		StartedAt:     time.Now(),
		EstimatedDays: 3,
	}

	for _, step := range OnboardingSteps {
		state.StepStatus[step.Key] = "pending"
	}

	if err := s.saveState(ctx, state); err != nil {
		return nil, err
	}

	// Ejecutar pasos automáticos en background
	go s.runAutoSteps(context.Background(), state)

	log.Printf("🚀 Onboarding iniciado: empresa %s — 9 pasos", companyID)
	return state, nil
}

func (s *OnboardingService) GetState(ctx context.Context, companyID string) (*OnboardingState, error) {
	var state OnboardingState
	result := s.db.WithContext(ctx).Table("onboarding_states").Where("company_id = ?", companyID).First(&state)
	if result.Error != nil {
		return nil, fmt.Errorf("estado de onboarding no encontrado")
	}
	return &state, nil
}

func (s *OnboardingService) CompleteStep(ctx context.Context, companyID, stepKey string) (*OnboardingState, error) {
	state, err := s.GetState(ctx, companyID)
	if err != nil {
		return nil, err
	}

	state.StepStatus[stepKey] = "done"

	// Avanzar al siguiente paso
	nextStep := s.findNextPendingStep(state)
	if nextStep == nil {
		// Onboarding completo
		now := time.Now()
		state.Completed = true
		state.CompletedAt = &now
		log.Printf("✅ Onboarding completado: empresa %s en %v", companyID, now.Sub(state.StartedAt).Round(time.Hour))
	} else {
		state.CurrentStep = nextStep.ID
		if nextStep.Auto {
			go s.executeStep(context.Background(), state, nextStep)
		}
	}

	return state, s.saveState(ctx, state)
}

func (s *OnboardingService) runAutoSteps(ctx context.Context, state *OnboardingState) {
	for _, step := range OnboardingSteps {
		if !step.Auto {
			continue
		}
		if state.StepStatus[step.Key] != "pending" {
			continue
		}

		if err := s.executeStep(ctx, state, &step); err != nil {
			log.Printf("❌ Paso automático %s falló: %v", step.Key, err)
			state.StepStatus[step.Key] = "failed"
		} else {
			state.StepStatus[step.Key] = "done"
		}
		s.saveState(ctx, state)
		time.Sleep(2 * time.Second) // pausa entre pasos
	}
}

func (s *OnboardingService) executeStep(ctx context.Context, state *OnboardingState, step *Step) error {
	state.StepStatus[step.Key] = "running"
	s.saveState(ctx, state)

	log.Printf("⚙️  Ejecutando paso %d: %s para empresa %s", step.ID, step.Title, state.CompanyID)

	// Publicar evento al servicio correspondiente
	payload, _ := json.Marshal(map[string]interface{}{
		"company_id": state.CompanyID,
		"tenant_id":  state.TenantID,
		"step_key":   step.Key,
		"step_id":    step.ID,
	})

	routingKey := fmt.Sprintf("onboarding.step.%s", step.Key)
	s.amqp.Publish("zu.companies", routingKey, payload)

	// Simular tiempo de procesamiento
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (s *OnboardingService) findNextPendingStep(state *OnboardingState) *Step {
	for i := range OnboardingSteps {
		step := &OnboardingSteps[i]
		if state.StepStatus[step.Key] == "pending" {
			return step
		}
	}
	return nil
}

func (s *OnboardingService) saveState(ctx context.Context, state *OnboardingState) error {
	return s.db.WithContext(ctx).Table("onboarding_states").
		Where("company_id = ?", state.CompanyID).
		Assign(state).
		FirstOrCreate(state).Error
}

func (s *OnboardingService) GetAllSteps() []Step { return OnboardingSteps }
