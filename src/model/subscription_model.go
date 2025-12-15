package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "ACTIVE"
	SubscriptionStatusInactive  SubscriptionStatus = "INACTIVE"
	SubscriptionStatusExpired   SubscriptionStatus = "EXPIRED"
	SubscriptionStatusSuspended SubscriptionStatus = "SUSPENDED"
)

type SubscriptionPlan string

const (
	SubscriptionPlanFree       SubscriptionPlan = "FREE"
	SubscriptionPlanBasic      SubscriptionPlan = "BASIC"
	SubscriptionPlanPro        SubscriptionPlan = "PRO"
	SubscriptionPlanEnterprise SubscriptionPlan = "ENTERPRISE"
)

type Subscription struct {
	ID            uuid.UUID          `gorm:"primaryKey;type:char(36)" json:"id"`
	UserID        uuid.UUID          `gorm:"not null;uniqueIndex" json:"userId"`
	User          *User              `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Plan          SubscriptionPlan   `gorm:"not null;default:FREE" json:"plan"`
	Status        SubscriptionStatus `gorm:"not null;default:ACTIVE" json:"status"`
	PointsBalance int                `gorm:"not null;default:0" json:"pointsBalance"`
	PointsUsed    int                `gorm:"not null;default:0" json:"pointsUsed"`
	PointsTotal   int                `gorm:"not null;default:0" json:"pointsTotal"`
	MonthlyPoints int                `gorm:"not null;default:0" json:"monthlyPoints"`
	LastResetDate time.Time          `gorm:"not null" json:"lastResetDate"`
	NextResetDate time.Time          `gorm:"not null" json:"nextResetDate"`
	CreatedAt     time.Time          `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time          `gorm:"autoUpdateTime" json:"updatedAt"`
	ExpiresAt     *time.Time         `gorm:"index" json:"expiresAt"`
}

func (subscription *Subscription) BeforeCreate(_ *gorm.DB) error {
	subscription.ID = uuid.New()
	return nil
}

// Utility functions for enum types

// IsValidSubscriptionPlan checks if the subscription plan is valid
func IsValidSubscriptionPlan(plan SubscriptionPlan) bool {
	switch plan {
	case SubscriptionPlanFree, SubscriptionPlanBasic,
		SubscriptionPlanPro, SubscriptionPlanEnterprise:
		return true
	default:
		return false
	}
}

// IsValidSubscriptionStatus checks if the subscription status is valid
func IsValidSubscriptionStatus(status SubscriptionStatus) bool {
	switch status {
	case SubscriptionStatusActive, SubscriptionStatusInactive,
		SubscriptionStatusExpired, SubscriptionStatusSuspended:
		return true
	default:
		return false
	}
}

// GetSubscriptionPlanDisplayName returns a human-readable name for the subscription plan
func GetSubscriptionPlanDisplayName(plan SubscriptionPlan) string {
	switch plan {
	case SubscriptionPlanFree:
		return "Free"
	case SubscriptionPlanBasic:
		return "Basic"
	case SubscriptionPlanPro:
		return "Pro"
	case SubscriptionPlanEnterprise:
		return "Enterprise"
	default:
		return "Unknown"
	}
}

// GetSubscriptionStatusDisplayName returns a human-readable name for the subscription status
func GetSubscriptionStatusDisplayName(status SubscriptionStatus) string {
	switch status {
	case SubscriptionStatusActive:
		return "Active"
	case SubscriptionStatusInactive:
		return "Inactive"
	case SubscriptionStatusExpired:
		return "Expired"
	case SubscriptionStatusSuspended:
		return "Suspended"
	default:
		return "Unknown"
	}
}

// GetSubscriptionPlanPoints returns the monthly points allocation for a plan
func GetSubscriptionPlanPoints(plan SubscriptionPlan) int {
	switch plan {
	case SubscriptionPlanFree:
		return 10
	case SubscriptionPlanBasic:
		return 500
	case SubscriptionPlanPro:
		return 1600
	case SubscriptionPlanEnterprise:
		return 5000
	default:
		return 0
	}
}

// GetPlanPrice returns the price for a subscription plan
func GetPlanPrice(plan SubscriptionPlan) float64 {
	switch plan {
	case SubscriptionPlanFree:
		return 0.0
	case SubscriptionPlanBasic:
		return 9.99
	case SubscriptionPlanPro:
		return 29.99
	case SubscriptionPlanEnterprise:
		return 99.99
	default:
		return 0.0
	}
}

// GetPointPrice returns the price per point for point purchases
func GetPointPrice(points int) float64 {
	// Bulk pricing: more points = better price
	switch {
	case points >= 1000:
		return 0.01 // $0.01 per point
	case points >= 500:
		return 0.015 // $0.015 per point
	case points >= 100:
		return 0.02 // $0.02 per point
	default:
		return 0.025 // $0.025 per point
	}
}

// IsSubscriptionActive checks if the subscription is active and not expired
func (sub *Subscription) IsSubscriptionActive() bool {
	if sub.Status != SubscriptionStatusActive {
		return false
	}
	if sub.ExpiresAt != nil && time.Now().After(*sub.ExpiresAt) {
		return false
	}
	return true
}

// HasEnoughPoints checks if user has enough points for a task
func (sub *Subscription) HasEnoughPoints(requiredPoints int) bool {
	return sub.PointsBalance >= requiredPoints
}

// CanPerformTask checks if user can perform a task based on subscription and points
func (sub *Subscription) CanPerformTask(requiredPoints int) bool {
	return sub.IsSubscriptionActive() && sub.HasEnoughPoints(requiredPoints)
}
