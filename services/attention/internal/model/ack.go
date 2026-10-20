package model

import "time"

// AttentionAck records that a tenant dismissed an inbox item. Inbox items
// themselves are virtual — computed live from incidents/medical_exams/
// documents in InboxService.GetPrioritizedInbox — so there is no pre-existing
// row to flip a flag on; acknowledging an item means inserting (or
// re-touching) its ack row here, and GetPrioritizedInbox excludes anything
// with one.
type AttentionAck struct {
	ItemID         string    `gorm:"primaryKey;size:64;column:item_id"`
	TenantID       string    `gorm:"primaryKey;size:64;column:tenant_id"`
	AcknowledgedAt time.Time `gorm:"column:acknowledged_at"`
}

func (AttentionAck) TableName() string { return "attention_acks" }
