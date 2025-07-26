package validation

import "github.com/mbbipul/tryonmagic-backend/src/model"

type SubscriptionPurchaseRequest struct {
	Plan          model.SubscriptionPlan `json:"plan" validate:"required,oneof=FREE BASIC PRO ENTERPRISE"`
	PaymentMethod model.PaymentMethod    `json:"paymentMethod" validate:"required,oneof=CREDIT_CARD PAYPAL STRIPE SHOPIFY"`
	ReferenceID   string                 `json:"referenceId" validate:"required"`
}
