package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

var ErrIdentityConflict = errors.New("account or company identifier already exists")

type UserRepository struct {
	db             *gorm.DB
	syntheticPilot bool
}

func NewUserRepository(db *gorm.DB, syntheticPilot ...bool) *UserRepository {
	var synthetic bool
	if len(syntheticPilot) > 0 {
		synthetic = syntheticPilot[0]
	}
	return &UserRepository{db: db, syntheticPilot: synthetic}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	query := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?) AND is_active = true", email)
	if r.syntheticPilot {
		query = query.Where("is_synthetic = true")
	}
	err := query.First(&u).Error
	return &u, err
}

func (r *UserRepository) FindByEmailAnyStatus(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	query := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?)", email)
	if r.syntheticPilot {
		query = query.Where("is_synthetic = true")
	}
	err := query.First(&u).Error
	return &u, err
}

func (r *UserRepository) FindByGoogleSubject(ctx context.Context, subject string) (*model.User, error) {
	var u model.User
	query := r.db.WithContext(ctx).Where("google_sub = ?", subject)
	if r.syntheticPilot {
		query = query.Where("is_synthetic = true")
	}
	err := query.First(&u).Error
	return &u, err
}

func (r *UserRepository) CompanyNITExists(ctx context.Context, nit string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Company{}).
		Where("regexp_replace(nit, '[^0-9]', '', 'g') = regexp_replace(?, '[^0-9]', '', 'g')", nit).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	query := r.db.WithContext(ctx).Where("id = ?", id)
	if r.syntheticPilot {
		query = query.Where("is_synthetic = true")
	}
	err := query.First(&u).Error
	return &u, err
}

func (r *UserRepository) CompanyBelongsToTenant(ctx context.Context, companyID, tenantID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Company{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", companyID, tenantID).
		Count(&count).Error
	return count == 1, err
}

func (r *UserRepository) HasActiveCompanyMembership(ctx context.Context, userID, tenantID, companyID uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Table("company_memberships AS m").
		Joins("JOIN companies AS c ON c.id = m.company_id AND c.tenant_id = m.tenant_id").
		Where("m.user_id = ? AND m.tenant_id = ? AND m.company_id = ? AND m.is_active = true AND c.deleted_at IS NULL", userID, tenantID, companyID)
	if r.syntheticPilot {
		query = query.Where("m.is_synthetic = true AND c.is_synthetic = true")
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count == 1, nil
}

func (r *UserRepository) ListMemberCompanies(ctx context.Context, userID, tenantID uuid.UUID) ([]model.Company, error) {
	var companies []model.Company
	query := r.db.WithContext(ctx).Table("companies AS c").
		Select("c.*").
		Joins("JOIN company_memberships AS m ON m.company_id = c.id AND m.tenant_id = c.tenant_id").
		Where("m.user_id = ? AND m.tenant_id = ? AND m.is_active = true AND c.deleted_at IS NULL", userID, tenantID)
	if r.syntheticPilot {
		query = query.Where("m.is_synthetic = true AND c.is_synthetic = true")
	}
	err := query.Order("c.name ASC").Find(&companies).Error
	return companies, err
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// CreateGoogleRegistration atomically provisions the first isolated workspace.
// It never creates demo data and deliberately leaves legal risk classification blank.
func (r *UserRepository) CreateGoogleRegistration(ctx context.Context, user *model.User, company *model.Company) error {
	if r.syntheticPilot {
		return errors.New("real Google registrations are disabled in synthetic mode")
	}
	if user == nil || company == nil || user.TenantID == uuid.Nil {
		return errors.New("invalid Google registration payload")
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if company.ID == uuid.Nil {
		company.ID = uuid.New()
	}
	company.TenantID = user.TenantID
	company.IsSynthetic = false
	company.RiskClass = ""
	company.RiskScore = 0
	company.EmployeeCount = 0
	company.PlanID = nil
	company.Status = model.CompanyStatusActive
	company.OnboardingStep = 1
	company.OnboardingDone = false
	user.CompanyID = &company.ID
	user.IsSynthetic = false
	user.IsActive = true

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrIdentityConflict
			}
			return err
		}
		if err := tx.Create(company).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrIdentityConflict
			}
			return err
		}
		membership := model.CompanyMembership{
			TenantID: user.TenantID, UserID: user.ID, CompanyID: company.ID,
			Role: "manager", IsActive: true, IsSynthetic: false,
		}
		if err := tx.Create(&membership).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrIdentityConflict
			}
			return err
		}
		return nil
	})
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("last_login", now).Error
}

func (r *UserRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND is_active = true", tenantID).Find(&users).Error
	return users, err
}
