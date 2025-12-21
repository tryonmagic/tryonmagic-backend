package validation

import "time"

// ShopifySyncRequest represents the request payload from Shopify for subscription sync
type ShopifySyncRequest struct {
	ShopDomain         string                 `json:"shopDomain" validate:"required"`
	Plan               string                 `json:"plan" validate:"required,oneof=FREE BASIC PRO ENTERPRISE"`
	Status             string                 `json:"status" validate:"required,oneof=ACTIVE INACTIVE EXPIRED SUSPENDED"`
	CurrentPeriodStart time.Time              `json:"currentPeriodStart" validate:"required"`
	CurrentPeriodEnd   time.Time              `json:"currentPeriodEnd" validate:"required"`
	Metadata           map[string]interface{} `json:"metadata"`
}
