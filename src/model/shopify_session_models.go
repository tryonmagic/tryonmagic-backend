package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShopifySession struct {
	ID               string                   `gorm:"primaryKey;not null"`
	AccessToken      *string                  `gorm:"null"`
	UserID           uuid.UUID                `gorm:"not null"`
	Expires          *time.Time               `gorm:"null"`
	IsOnline         bool                     `gorm:"default:false;not null"`
	Scope            *string                  `gorm:"null"`
	State            string                   `gorm:"null"`
	APIKey           string                   `gorm:"not null"`
	Shop             string                   `gorm:"not null"`
	CreatedAt        time.Time                `gorm:"autoCreateTime:milli"`
	UpdatedAt        time.Time                `gorm:"autoCreateTime:milli;autoUpdateTime:milli"`
	User             *User                    `gorm:"foreignKey:UserID;references:ID"`
	OnlineAccessInfo *ShopifyOnlineAccessInfo `gorm:"foreignKey:ShopifySessionID;references:ID"`
}

type ShopifyOnlineAccessInfo struct {
	ID                  uuid.UUID              `gorm:"primaryKey;not null"`
	ShopifySessionID    string                 `gorm:"unique,index,not null,"`
	ExpiresIn           int                    `gorm:"not null"`
	AssociatedUserScope string                 `gorm:"not null"`
	CreatedAt           time.Time              `gorm:"autoCreateTime:milli"`
	UpdatedAt           time.Time              `gorm:"autoCreateTime:milli;autoUpdateTime:milli"`
	AssociatedUser      *ShopifyAssociatedUser `gorm:"foreignKey:OnlineAccessInfoID;references:id"`
	ShopifySession      *ShopifySession        `gorm:"foreignKey:ShopifySessionID;references:id"`
}

type ShopifyAssociatedUser struct {
	ID                 uuid.UUID                `gorm:"primaryKey;not null"`
	OnlineAccessInfoID *uuid.UUID               `gorm:"unique"`
	ShopifyUserID      int64                    `gorm:"not null"`
	FirstName          string                   `gorm:"not null"`
	LastName           string                   `gorm:"not null"`
	Email              string                   `gorm:"not null"`
	AccountOwner       bool                     `gorm:"not null"`
	Locale             string                   `gorm:"not null"`
	Collaborator       bool                     `gorm:"not null"`
	EmailVerified      bool                     `gorm:"not null"`
	CreatedAt          time.Time                `gorm:"autoCreateTime:milli"`
	UpdatedAt          time.Time                `gorm:"autoUpdateTime:milli"`
	OnlineAccessInfo   *ShopifyOnlineAccessInfo `gorm:"foreignKey:OnlineAccessInfoID;references:id"`
}

func (session *ShopifyOnlineAccessInfo) BeforeCreate(_ *gorm.DB) error {
	session.ID = uuid.New()
	return nil
}

func (session *ShopifySession) BeforeCreate(_ *gorm.DB) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	return nil
}

func (session *ShopifyAssociatedUser) BeforeCreate(_ *gorm.DB) error {
	session.ID = uuid.New()
	return nil
}
