package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleUser            Role = "USER"
	RoleAdmin           Role = "ADMIN"
	RoleInternalService Role = "INTERNAL_SERVICE"
)

type User struct {
	ID            uuid.UUID `gorm:"primaryKey;not null" json:"id"`
	FirstName     string    `gorm:"not null" json:"firstName"`
	LastName      string    `gorm:"not null" json:"lastName"`
	Email         string    `gorm:"not null" json:"email"`
	Password      string    `gorm:"not null" json:"-"`
	Role          Role      `gorm:"default:USER;not null" json:"role"`
	EmailVerified bool      `gorm:"default:false;not null" json:"emailVerified"`
	// Shopify-specific fields (nullable for non-Shopify users)
	Shop          *string       `gorm:"index:idx_user_shop_shopify_user_id" json:"shop,omitempty"`
	ShopifyUserID *int64        `gorm:"index:idx_user_shop_shopify_user_id" json:"shopifyUserID,omitempty"`
	CreatedAt     time.Time     `gorm:"autoCreateTime:milli" json:"-"`
	UpdatedAt     time.Time     `gorm:"autoCreateTime:milli;autoUpdateTime:milli" json:"-"`
	Token         []Token       `gorm:"foreignKey:user_id;references:id" json:"-"`
	Subscription  *Subscription `gorm:"foreignKey:user_id;references:id" json:"subscription"`
}

func (user *User) BeforeCreate(_ *gorm.DB) error {
	user.ID = uuid.New() // Generate UUID before create
	return nil
}
