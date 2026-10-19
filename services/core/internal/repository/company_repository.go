package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type CompanyRepository struct {
	db             *gorm.DB
	syntheticPilot bool
}

func NewCompanyRepository(db *gorm.DB, syntheticPilot ...bool) *CompanyRepository {
	var synthetic bool
	if len(syntheticPilot) > 0 {
		synthetic = syntheticPilot[0]
	}
	return &CompanyRepository{db: db, syntheticPilot: synthetic}
}

func (r *CompanyRepository) FindAll(ctx context.Context, tenantID uuid.UUID) ([]model.Company, error) {
	var companies []model.Company
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("risk_score DESC").
		Find(&companies).Error
	return companies, err
}

func (r *CompanyRepository) FindAllForUser(ctx context.Context, userID, tenantID uuid.UUID) ([]model.Company, error) {
	var companies []model.Company
	query := r.db.WithContext(ctx).Table("companies AS c").
		Select("c.*").
		Joins("JOIN company_memberships AS m ON m.company_id = c.id AND m.tenant_id = c.tenant_id").
		Where("m.user_id = ? AND m.tenant_id = ? AND m.is_active = true AND c.deleted_at IS NULL", userID, tenantID)
	if r.syntheticPilot {
		query = query.Where("m.is_synthetic = true AND c.is_synthetic = true")
	}
	err := query.Order("c.risk_score DESC, c.name ASC").Find(&companies).Error
	return companies, err
}

func (r *CompanyRepository) CreateForUser(ctx context.Context, company *model.Company, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(company).Error; err != nil {
			return err
		}
		membership := model.CompanyMembership{
			TenantID: company.TenantID, UserID: userID, CompanyID: company.ID,
			Role: "manager", IsActive: true, IsSynthetic: company.IsSynthetic,
		}
		return tx.Create(&membership).Error
	})
}

func (r *CompanyRepository) FindByID(ctx context.Context, id, tenantID uuid.UUID) (*model.Company, error) {
	var c model.Company
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&c).Error
	return &c, err
}

func (r *CompanyRepository) FindByNIT(ctx context.Context, nit string) (*model.Company, error) {
	var c model.Company
	err := r.db.WithContext(ctx).Where("nit = ?", nit).First(&c).Error
	return &c, err
}

func (r *CompanyRepository) Create(ctx context.Context, c *model.Company) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CompanyRepository) Update(ctx context.Context, c *model.Company) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CompanyRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&model.Company{}).Error
}

func (r *CompanyRepository) UpdateRiskScore(ctx context.Context, id uuid.UUID, score float64) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Where("id = ?", id).
		Update("risk_score", score).Error
}

func (r *CompanyRepository) UpdateOnboardingStep(ctx context.Context, id uuid.UUID, step int, done bool) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"onboarding_step": step, "onboarding_done": done}).Error
}

func (r *CompanyRepository) CountByCIIU(ctx context.Context, ciiu string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Company{}).Where("ciiu = ?", ciiu).Count(&count).Error
	return count, err
}

func (r *CompanyRepository) FindCritical(ctx context.Context, tenantID uuid.UUID) ([]model.Company, error) {
	var companies []model.Company
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND risk_score < 65 AND status = 'active'", tenantID).
		Order("risk_score ASC").Find(&companies).Error
	return companies, err
}
