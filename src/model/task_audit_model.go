package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditStatus string

const (
	AuditStatusPending  AuditStatus = "PENDING"
	AuditStatusResolved AuditStatus = "RESOLVED"
	AuditStatusRejected AuditStatus = "REJECTED"
)

type TaskAudit struct {
	ID                uuid.UUID   `gorm:"primaryKey;type:char(36)" json:"id"`
	TaskID            uuid.UUID   `gorm:"not null;index" json:"taskId"`
	Task              *Task       `gorm:"foreignKey:TaskID;references:ID" json:"task"`
	UserID            uuid.UUID   `gorm:"not null;index" json:"userId"`
	User              *User       `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Reason            string      `gorm:"type:text;not null" json:"reason"`
	IsRefundRequested bool        `gorm:"not null;default:false" json:"isRefundRequested"`
	AdminFeedback     string      `gorm:"type:text" json:"adminFeedback"`
	Status            AuditStatus `gorm:"type:varchar(20);not null;default:PENDING" json:"status"`
	CreatedAt         time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ta *TaskAudit) BeforeCreate(tx *gorm.DB) error {
	ta.ID = uuid.New()
	return nil
}
