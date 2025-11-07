package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GDPR Request Types
type GDPRRequestType string

const (
	GDPRRequestTypeDataRequest GDPRRequestType = "data_request"
	GDPRRequestTypeRedact      GDPRRequestType = "redact"
	GDPRRequestTypeShopRedact  GDPRRequestType = "shop_redact"
)

// GDPR Request Status
type GDPRRequestStatus string

const (
	GDPRRequestStatusPending    GDPRRequestStatus = "pending"
	GDPRRequestStatusProcessing GDPRRequestStatus = "processing"
	GDPRRequestStatusCompleted  GDPRRequestStatus = "completed"
	GDPRRequestStatusFailed     GDPRRequestStatus = "failed"
)

// GDPRRequest represents a GDPR compliance request log
type GDPRRequest struct {
	ID            uuid.UUID         `gorm:"primaryKey;type:char(36)" json:"id"`
	Shop          string            `gorm:"not null;index" json:"shop"`
	CustomerID    string            `gorm:"not null;index" json:"customerId"`
	CustomerEmail string            `gorm:"not null" json:"customerEmail"`
	RequestType   GDPRRequestType   `gorm:"not null" json:"requestType"`
	Status        GDPRRequestStatus `gorm:"default:pending;not null" json:"status"`
	Payload       datatypes.JSON    `gorm:"type:text" json:"payload"`
	Notes         *string           `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time         `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime" json:"updatedAt"`
	ProcessedAt   *time.Time        `json:"processedAt"`
}

// AppInstallation represents a Shopify app installation record
type AppInstallation struct {
	ID            uuid.UUID  `gorm:"primaryKey;type:char(36)" json:"id"`
	Shop          string     `gorm:"uniqueIndex;not null" json:"shop"`
	AccessToken   string     `gorm:"not null" json:"accessToken"`
	Scope         string     `gorm:"not null" json:"scope"`
	IsActive      bool       `gorm:"default:true;not null" json:"isActive"`
	InstalledAt   time.Time  `gorm:"autoCreateTime" json:"installedAt"`
	UninstalledAt *time.Time `json:"uninstalledAt"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// UserPreferences represents customer settings and preferences
type UserPreferences struct {
	ID          uuid.UUID      `gorm:"primaryKey;type:char(36)" json:"id"`
	Shop        string         `gorm:"not null;index" json:"shop"`
	CustomerID  string         `gorm:"not null;index" json:"customerId"`
	Preferences datatypes.JSON `gorm:"type:text" json:"preferences"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

// GeneratedImage represents AI-generated content for GDPR tracking
type GeneratedImage struct {
	ID         uuid.UUID      `gorm:"primaryKey;type:char(36)" json:"id"`
	Shop       string         `gorm:"not null;index" json:"shop"`
	CustomerID string         `gorm:"not null;index" json:"customerId"`
	ImageURL   string         `gorm:"not null" json:"imageUrl"`
	Metadata   datatypes.JSON `gorm:"type:text" json:"metadata"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"createdAt"`
}

// UsageHistory represents analytics data for GDPR compliance
type UsageHistory struct {
	ID         uuid.UUID      `gorm:"primaryKey;type:char(36)" json:"id"`
	Shop       string         `gorm:"not null;index" json:"shop"`
	CustomerID string         `gorm:"not null;index" json:"customerId"`
	Action     string         `gorm:"not null" json:"action"`
	Data       datatypes.JSON `gorm:"type:text" json:"data"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"createdAt"`
}

// ShopSettings represents shop configuration data
type ShopSettings struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:char(36)" json:"id"`
	Shop      string         `gorm:"uniqueIndex;not null" json:"shop"`
	Settings  datatypes.JSON `gorm:"type:text" json:"settings"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

// ShopDeletion represents deletion audit log
type ShopDeletion struct {
	ID          uuid.UUID `gorm:"primaryKey;type:char(36)" json:"id"`
	Shop        string    `gorm:"not null;index" json:"shop"`
	DeletedAt   time.Time `gorm:"autoCreateTime" json:"deletedAt"`
	Reason      string    `gorm:"not null" json:"reason"`
	DataDeleted int       `gorm:"default:0" json:"dataDeleted"`
}

// BeforeCreate hooks for UUID generation
func (g *GDPRRequest) BeforeCreate(_ *gorm.DB) error {
	g.ID = uuid.New()
	return nil
}

func (a *AppInstallation) BeforeCreate(_ *gorm.DB) error {
	a.ID = uuid.New()
	return nil
}

func (u *UserPreferences) BeforeCreate(_ *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

func (g *GeneratedImage) BeforeCreate(_ *gorm.DB) error {
	g.ID = uuid.New()
	return nil
}

func (u *UsageHistory) BeforeCreate(_ *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

func (s *ShopSettings) BeforeCreate(_ *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}

func (s *ShopDeletion) BeforeCreate(_ *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}
