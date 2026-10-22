// Package store persists in-app notifications and an audit row per delivery
// attempt, in the service's own schema (search_path set on its DB role).
package store

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  string     `gorm:"type:uuid;not null;index:idx_notif_scope,priority:1" json:"tenant_id"`
	UserID    *string    `gorm:"type:uuid;index:idx_notif_scope,priority:2" json:"user_id,omitempty"` // nil = every user of the tenant
	CompanyID *string    `gorm:"type:uuid" json:"company_id,omitempty"`
	Kind      string     `gorm:"size:64;not null" json:"kind"`
	Title     string     `gorm:"size:255;not null" json:"title"`
	Message   string     `gorm:"type:text;not null" json:"message"`
	Priority  string     `gorm:"size:16;not null;default:normal" json:"priority"`
	Link      string     `gorm:"size:512" json:"link,omitempty"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
}

type Delivery struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	NotificationID *string   `gorm:"type:uuid;index"`
	TenantID       string    `gorm:"type:uuid;not null;index"`
	Channel        string    `gorm:"size:16;not null"`
	Kind           string    `gorm:"size:64;not null"`
	Success        bool      `gorm:"not null"`
	Error          string    `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"index"`
}

type Store struct{ db *gorm.DB }

func New(db *gorm.DB) *Store { return &Store{db: db} }

func (s *Store) Migrate() error { return s.db.AutoMigrate(&Notification{}, &Delivery{}) }

func (s *Store) Create(ctx context.Context, n *Notification) error {
	return s.db.WithContext(ctx).Create(n).Error
}

func (s *Store) RecordDelivery(ctx context.Context, d *Delivery) error {
	return s.db.WithContext(ctx).Create(d).Error
}

// visible scopes every read and write to what the caller may see: its own
// tenant, and either addressed to it or tenant-wide.
func (s *Store) visible(ctx context.Context, tenantID, userID string) *gorm.DB {
	return s.db.WithContext(ctx).Model(&Notification{}).
		Where("tenant_id = ? AND (user_id = ? OR user_id IS NULL)", tenantID, userID)
}

func (s *Store) List(ctx context.Context, tenantID, userID string, unreadOnly bool, limit int) ([]Notification, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var unread int64
	if err := s.visible(ctx, tenantID, userID).Where("read_at IS NULL").Count(&unread).Error; err != nil {
		return nil, 0, err
	}
	q := s.visible(ctx, tenantID, userID)
	if unreadOnly {
		q = q.Where("read_at IS NULL")
	}
	var out []Notification
	err := q.Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, unread, err
}

// MarkRead returns gorm.ErrRecordNotFound when the id is not visible to the
// caller, so foreign ids are indistinguishable from missing ones.
func (s *Store) MarkRead(ctx context.Context, tenantID, userID, id string) error {
	res := s.visible(ctx, tenantID, userID).Where("id = ? AND read_at IS NULL", id).Update("read_at", time.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var n int64
		s.visible(ctx, tenantID, userID).Where("id = ?", id).Count(&n)
		if n == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (s *Store) MarkAllRead(ctx context.Context, tenantID, userID string) (int64, error) {
	res := s.visible(ctx, tenantID, userID).Where("read_at IS NULL").Update("read_at", time.Now())
	return res.RowsAffected, res.Error
}
