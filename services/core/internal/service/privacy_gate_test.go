package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestConsentServicesFailClosedWithoutDatabaseAccess(t *testing.T) {
	ctx := context.Background()
	tenantID, employeeID, recordID := uuid.New(), uuid.New(), uuid.New()

	consent := NewConsentService(nil)
	if record, err := consent.GrantConsent(ctx, tenantID, employeeID, "salud", "anything", "web", "https://example.test/evidence"); record != nil || !errors.Is(err, ErrConsentCaptureDisabled) {
		t.Fatalf("GrantConsent must fail closed; record=%v err=%v", record, err)
	}
	if valid, err := consent.HasValidConsent(ctx, tenantID, employeeID, "salud"); valid || !errors.Is(err, ErrConsentAuthorizationUnavailable) {
		t.Fatalf("HasValidConsent must deny unverifiable records; valid=%v err=%v", valid, err)
	}
	if err := consent.RevokeConsent(ctx, tenantID, recordID); !errors.Is(err, ErrConsentManagementDisabled) {
		t.Fatalf("RevokeConsent must require actor-attributed audit; err=%v", err)
	}
	if _, err := consent.RequestDataErasure(ctx, tenantID, employeeID, recordID); !errors.Is(err, ErrDataErasureDisabled) {
		t.Fatalf("RequestDataErasure must remain closed; err=%v", err)
	}

	privacy := NewPrivacyService(nil)
	if record, err := privacy.RegisterConsent(ctx, tenantID, employeeID, "127.0.0.1", "some evidence"); record != nil || !errors.Is(err, ErrConsentCaptureDisabled) {
		t.Fatalf("RegisterConsent must not fabricate consent; record=%v err=%v", record, err)
	}
	if valid, err := privacy.HasValidConsent(ctx, tenantID, employeeID); valid || !errors.Is(err, ErrConsentAuthorizationUnavailable) {
		t.Fatalf("PrivacyService.HasValidConsent must fail closed; valid=%v err=%v", valid, err)
	}
	if err := privacy.RevokeConsent(ctx, tenantID, employeeID); !errors.Is(err, ErrConsentManagementDisabled) {
		t.Fatalf("PrivacyService.RevokeConsent must require actor-attributed audit; err=%v", err)
	}
	if err := privacy.EraseEmployeeData(ctx, tenantID, employeeID, "request"); !errors.Is(err, ErrPrivacyErasureDisabled) {
		t.Fatalf("EraseEmployeeData must remain closed; err=%v", err)
	}
}

func TestMedicalExamServiceBlocksAllOperationsBeforeDependencies(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	svc := NewMedicalExamService(nil, nil)

	if exam, err := svc.Schedule(ctx, ScheduleExamInput{}, id, id); exam != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("Schedule must fail closed; exam=%v err=%v", exam, err)
	}
	if err := svc.ProcessOCR(ctx, id, map[string]interface{}{"raw": "medical data"}); !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("ProcessOCR must fail closed; err=%v", err)
	}
	if exam, err := svc.Review(ctx, id, id, id, ReviewExamInput{}); exam != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("Review must fail closed; exam=%v err=%v", exam, err)
	}
	if exams, err := svc.GetByEmployee(ctx, id, id); exams != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GetByEmployee must fail closed; exams=%v err=%v", exams, err)
	}
	if exams, err := svc.GetPendingReview(ctx, id); exams != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GetPendingReview must fail closed; exams=%v err=%v", exams, err)
	}
	if exams, err := svc.GetExpiring(ctx, id, 30); exams != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GetExpiring must fail closed; exams=%v err=%v", exams, err)
	}
}

func TestDataExportServiceBlocksBeforeDatabaseAccess(t *testing.T) {
	ctx := context.Background()
	svc := NewDataExportService(nil)
	id := uuid.New()
	if export, err := svc.ExportCompany(ctx, id, id); export != nil || !errors.Is(err, ErrSensitiveDataExportDisabled) {
		t.Fatalf("ExportCompany must fail closed; export=%v err=%v", export, err)
	}
	if content, err := svc.ExportCompanyJSON(ctx, id, id); content != nil || !errors.Is(err, ErrSensitiveDataExportDisabled) {
		t.Fatalf("ExportCompanyJSON must fail closed; content=%v err=%v", content, err)
	}
}

func TestTrainingServiceBlocksWorkerDataAndCertificateOperations(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	svc := NewTrainingService(nil, nil)

	if training, err := svc.Create(ctx, CreateTrainingInput{}, id, id); training != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("Create must fail closed; training=%v err=%v", training, err)
	}
	if trainings, err := svc.List(ctx, id, id); trainings != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("List must fail closed; trainings=%v err=%v", trainings, err)
	}
	if certs, err := svc.Complete(ctx, id, id, id, RegisterAttendanceInput{}); certs != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("Complete must fail closed; certs=%v err=%v", certs, err)
	}
	if certs, err := svc.GetExpiringCertificates(ctx, id, &id, 30); certs != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GetExpiringCertificates must fail closed; certs=%v err=%v", certs, err)
	}
	if cert, err := svc.VerifyCertificate(ctx, "synthetic-folio"); cert != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("VerifyCertificate must fail closed; cert=%v err=%v", cert, err)
	}
}

func TestEmployeeServiceBlocksWorkerDataOperations(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	svc := NewEmployeeService(nil, nil)

	if employees, total, err := svc.List(ctx, id, id, 1, 10); employees != nil || total != 0 || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("List must fail closed; employees=%v total=%d err=%v", employees, total, err)
	}
	if employee, err := svc.GetByID(ctx, id, id, id); employee != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GetByID must fail closed; employee=%v err=%v", employee, err)
	}
	if employee, err := svc.Create(ctx, CreateEmployeeInput{}, id, id); employee != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("Create must fail closed; employee=%v err=%v", employee, err)
	}
	if count, rowErrors, err := svc.BulkImportFromCSV(ctx, nil, id, id); count != 0 || rowErrors != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("BulkImportFromCSV must fail closed; count=%d rowErrors=%v err=%v", count, rowErrors, err)
	}
	if employee, err := svc.Update(ctx, id, id, id, UpdateEmployeeInput{}); employee != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("Update must fail closed; employee=%v err=%v", employee, err)
	}
	if err := svc.Delete(ctx, id, id, id); !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("Delete must fail closed; err=%v", err)
	}
}

func TestIncidentServiceBlocksIncidentAndFURATOperations(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	svc := NewIncidentService(nil, nil)

	if incident, err := svc.Create(ctx, CreateIncidentInput{}, id, id, id); incident != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("Create must fail closed; incident=%v err=%v", incident, err)
	}
	if incidents, err := svc.GetByCompany(ctx, id, id); incidents != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GetByCompany must fail closed; incidents=%v err=%v", incidents, err)
	}
	if incident, err := svc.GetByID(ctx, id, id, id); incident != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GetByID must fail closed; incident=%v err=%v", incident, err)
	}
	if incidents, err := svc.GetOpenByTenant(ctx, id, &id); incidents != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GetOpenByTenant must fail closed; incidents=%v err=%v", incidents, err)
	}
	if incident, err := svc.UpdateInvestigation(ctx, id, id, id, id, UpdateInvestigationInput{}); incident != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("UpdateInvestigation must fail closed; incident=%v err=%v", incident, err)
	}
	if furat, err := svc.GenerateFURAT(ctx, id, id, id); furat != nil || !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("GenerateFURAT must fail closed; furat=%v err=%v", furat, err)
	}
}
