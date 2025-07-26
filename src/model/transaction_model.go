package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeTaskExecution        TransactionType = "TASK_EXECUTION"        // Points used for task execution
	TransactionTypeSubscriptionPurchase TransactionType = "SUBSCRIPTION_PURCHASE" // Buying subscription plan
	TransactionTypePointPurchase        TransactionType = "POINT_PURCHASE"        // Buying additional points
	TransactionTypeRefund               TransactionType = "REFUND"                // Refund for failed task
	TransactionTypeBonus                TransactionType = "BONUS"                 // Bonus points from admin
	TransactionTypePlanUpgrade          TransactionType = "PLAN_UPGRADE"          // Upgrading subscription plan
	TransactionTypePlanDowngrade        TransactionType = "PLAN_DOWNGRADE"        // Downgrading subscription plan
	TransactionTypeMonthlyReset         TransactionType = "MONTHLY_RESET"         // Monthly points reset
	TransactionTypeReferralBonus        TransactionType = "REFERRAL_BONUS"        // Referral bonus
	TransactionTypePromotionalCredit    TransactionType = "PROMOTIONAL_CREDIT"    // Promotional credits
	TransactionTypeCompensation         TransactionType = "COMPENSATION"          // Compensation for issues
	TransactionTypeAdjustment           TransactionType = "ADJUSTMENT"            // Manual adjustment by admin
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

// PaymentMethod represents the method of payment
type PaymentMethod string

const (
	PaymentMethodCreditCard   PaymentMethod = "CREDIT_CARD"
	PaymentMethodPayPal       PaymentMethod = "PAYPAL"
	PaymentMethodStripe       PaymentMethod = "STRIPE"
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodFree         PaymentMethod = "FREE"
	PaymentMethodCrypto       PaymentMethod = "CRYPTO"
	// Partner payment methods
	PaymentMethodShopify PaymentMethod = "SHOPIFY"
	PaymentMethodPartner PaymentMethod = "PARTNER" // Generic partner payment method
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
)

// ReferenceType represents the type of reference
type ReferenceType string

const (
	ReferenceTypePayment      ReferenceType = "PAYMENT"
	ReferenceTypeOrder        ReferenceType = "ORDER"
	ReferenceTypeTask         ReferenceType = "TASK"
	ReferenceTypeSubscription ReferenceType = "SUBSCRIPTION"
	ReferenceTypeRefund       ReferenceType = "REFUND"
	ReferenceTypeBonus        ReferenceType = "BONUS"
)

// TransactionHistory represents a comprehensive transaction record
type TransactionHistory struct {
	ID                uuid.UUID         `gorm:"primaryKey;type:char(36)" json:"id"`
	UserID            uuid.UUID         `gorm:"not null;index" json:"userId"`
	User              *User             `gorm:"foreignKey:UserID;references:ID" json:"user"`
	TaskID            *uuid.UUID        `gorm:"index" json:"taskId"`
	Task              *Task             `gorm:"foreignKey:TaskID;references:ID" json:"task"`
	TransactionType   TransactionType   `gorm:"not null;index" json:"transactionType"`               // TASK_EXECUTION, SUBSCRIPTION_PURCHASE, POINT_PURCHASE, REFUND, BONUS, PLAN_UPGRADE, PLAN_DOWNGRADE, MONTHLY_RESET
	TransactionStatus TransactionStatus `gorm:"not null;default:COMPLETED" json:"transactionStatus"` // PENDING, COMPLETED, FAILED, CANCELLED
	DebitAmount       int               `gorm:"not null;default:0" json:"debitAmount"`               // Points deducted
	CreditAmount      int               `gorm:"not null;default:0" json:"creditAmount"`              // Points added
	NetAmount         int               `gorm:"not null" json:"netAmount"`                           // Credit - Debit
	BalanceBefore     int               `gorm:"not null" json:"balanceBefore"`
	BalanceAfter      int               `gorm:"not null" json:"balanceAfter"`
	Description       string            `gorm:"not null" json:"description"`
	ReferenceID       string            `gorm:"index" json:"referenceId"`   // External reference (payment ID, order ID, etc.)
	ReferenceType     ReferenceType     `gorm:"index" json:"referenceType"` // PAYMENT, ORDER, TASK, SUBSCRIPTION
	PaymentMethod     PaymentMethod     `gorm:"index" json:"paymentMethod"` // CREDIT_CARD, PAYPAL, STRIPE, BANK_TRANSFER, FREE, SHOPIFY, WOOCOMMERCE, etc.
	PaymentStatus     PaymentStatus     `gorm:"index" json:"paymentStatus"` // PENDING, COMPLETED, FAILED, REFUNDED
	Currency          string            `gorm:"default:USD" json:"currency"`
	AmountPaid        float64           `gorm:"default:0" json:"amountPaid"`   // Actual money paid
	ExchangeRate      float64           `gorm:"default:1" json:"exchangeRate"` // Points per dollar
	Metadata          string            `gorm:"type:text" json:"metadata"`     // JSON string for additional data
	CreatedAt         time.Time         `gorm:"autoCreateTime;index" json:"createdAt"`
	UpdatedAt         time.Time         `gorm:"autoUpdateTime" json:"updatedAt"`
	ProcessedAt       *time.Time        `gorm:"index" json:"processedAt"`
	ExpiresAt         *time.Time        `gorm:"index" json:"expiresAt"`
}

func (th *TransactionHistory) BeforeCreate(tx *gorm.DB) error {
	th.ID = uuid.New()
	if th.NetAmount == 0 {
		th.NetAmount = th.CreditAmount - th.DebitAmount
	}
	if th.ProcessedAt == nil {
		now := time.Now()
		th.ProcessedAt = &now
	}
	return nil
}

// GetTransactionSummary returns a summary of transaction statistics
func (th *TransactionHistory) GetTransactionSummary() map[string]interface{} {
	return map[string]interface{}{
		"transactionType":   th.TransactionType,
		"transactionStatus": th.TransactionStatus,
		"debitAmount":       th.DebitAmount,
		"creditAmount":      th.CreditAmount,
		"netAmount":         th.NetAmount,
		"balanceBefore":     th.BalanceBefore,
		"balanceAfter":      th.BalanceAfter,
		"paymentMethod":     th.PaymentMethod,
		"paymentStatus":     th.PaymentStatus,
		"amountPaid":        th.AmountPaid,
		"currency":          th.Currency,
	}
}

// IsDebitTransaction checks if this is a debit transaction (points deducted)
func (th *TransactionHistory) IsDebitTransaction() bool {
	return th.DebitAmount > 0 && th.CreditAmount == 0
}

// IsCreditTransaction checks if this is a credit transaction (points added)
func (th *TransactionHistory) IsCreditTransaction() bool {
	return th.CreditAmount > 0 && th.DebitAmount == 0
}

// IsPaymentTransaction checks if this transaction involves payment
func (th *TransactionHistory) IsPaymentTransaction() bool {
	return th.AmountPaid > 0 && th.PaymentMethod != PaymentMethodFree
}

// GetTransactionCategory returns the category of the transaction
func (th *TransactionHistory) GetTransactionCategory() string {
	switch th.TransactionType {
	case TransactionTypeTaskExecution:
		return "TASK_USAGE"
	case TransactionTypeSubscriptionPurchase, TransactionTypePointPurchase:
		return "PURCHASE"
	case TransactionTypeRefund, TransactionTypeCompensation:
		return "REFUND"
	case TransactionTypeBonus, TransactionTypeReferralBonus, TransactionTypePromotionalCredit:
		return "BONUS"
	case TransactionTypePlanUpgrade, TransactionTypePlanDowngrade:
		return "PLAN_CHANGE"
	case TransactionTypeMonthlyReset:
		return "RESET"
	case TransactionTypeAdjustment:
		return "ADMIN_ADJUSTMENT"
	default:
		return "OTHER"
	}
}

// Utility functions for enum types

// IsValidTransactionType checks if the transaction type is valid
func IsValidTransactionType(transactionType TransactionType) bool {
	switch transactionType {
	case TransactionTypeTaskExecution, TransactionTypeSubscriptionPurchase,
		TransactionTypePointPurchase, TransactionTypeRefund, TransactionTypeBonus,
		TransactionTypePlanUpgrade, TransactionTypePlanDowngrade, TransactionTypeMonthlyReset,
		TransactionTypeReferralBonus, TransactionTypePromotionalCredit,
		TransactionTypeCompensation, TransactionTypeAdjustment:
		return true
	default:
		return false
	}
}

// IsValidTransactionStatus checks if the transaction status is valid
func IsValidTransactionStatus(status TransactionStatus) bool {
	switch status {
	case TransactionStatusPending, TransactionStatusCompleted,
		TransactionStatusFailed, TransactionStatusCancelled:
		return true
	default:
		return false
	}
}

// IsValidPaymentMethod checks if the payment method is valid
func IsValidPaymentMethod(method PaymentMethod) bool {
	switch method {
	case PaymentMethodCreditCard, PaymentMethodPayPal, PaymentMethodStripe,
		PaymentMethodBankTransfer, PaymentMethodFree, PaymentMethodCrypto,
		PaymentMethodShopify, PaymentMethodPartner:
		return true
	default:
		return false
	}
}

// IsValidPaymentStatus checks if the payment status is valid
func IsValidPaymentStatus(status PaymentStatus) bool {
	switch status {
	case PaymentStatusPending, PaymentStatusCompleted, PaymentStatusFailed,
		PaymentStatusRefunded, PaymentStatusCancelled:
		return true
	default:
		return false
	}
}

// IsValidReferenceType checks if the reference type is valid
func IsValidReferenceType(refType ReferenceType) bool {
	switch refType {
	case ReferenceTypePayment, ReferenceTypeOrder, ReferenceTypeTask,
		ReferenceTypeSubscription, ReferenceTypeRefund, ReferenceTypeBonus:
		return true
	default:
		return false
	}
}

// GetTransactionTypeDisplayName returns a human-readable name for the transaction type
func GetTransactionTypeDisplayName(transactionType TransactionType) string {
	switch transactionType {
	case TransactionTypeTaskExecution:
		return "Task Execution"
	case TransactionTypeSubscriptionPurchase:
		return "Subscription Purchase"
	case TransactionTypePointPurchase:
		return "Point Purchase"
	case TransactionTypeRefund:
		return "Refund"
	case TransactionTypeBonus:
		return "Bonus"
	case TransactionTypePlanUpgrade:
		return "Plan Upgrade"
	case TransactionTypePlanDowngrade:
		return "Plan Downgrade"
	case TransactionTypeMonthlyReset:
		return "Monthly Reset"
	case TransactionTypeReferralBonus:
		return "Referral Bonus"
	case TransactionTypePromotionalCredit:
		return "Promotional Credit"
	case TransactionTypeCompensation:
		return "Compensation"
	case TransactionTypeAdjustment:
		return "Adjustment"
	default:
		return "Unknown"
	}
}

// GetPaymentMethodDisplayName returns a human-readable name for the payment method
func GetPaymentMethodDisplayName(method PaymentMethod) string {
	switch method {
	case PaymentMethodCreditCard:
		return "Credit Card"
	case PaymentMethodPayPal:
		return "PayPal"
	case PaymentMethodStripe:
		return "Stripe"
	case PaymentMethodBankTransfer:
		return "Bank Transfer"
	case PaymentMethodFree:
		return "Free"
	case PaymentMethodCrypto:
		return "Cryptocurrency"
	case PaymentMethodShopify:
		return "Shopify"

	case PaymentMethodPartner:
		return "Partner"
	default:
		return "Unknown"
	}
}

// GetPaymentMethodCategory returns the category of the payment method
func GetPaymentMethodCategory(method PaymentMethod) string {
	switch method {
	case PaymentMethodCreditCard:
		return "CREDIT_CARD"
	case PaymentMethodPayPal:
		return "DIGITAL_WALLET"
	case PaymentMethodStripe:
		return "DIGITAL_WALLET"
	case PaymentMethodShopify:
		return "ECOMMERCE_PLATFORM"
	case PaymentMethodCrypto:
		return "CRYPTOCURRENCY"
	case PaymentMethodBankTransfer:
		return "BANK_TRANSFER"
	case PaymentMethodFree:
		return "FREE"
	case PaymentMethodPartner:
		return "PARTNER"
	default:
		return "OTHER"
	}
}

// IsPartnerPaymentMethod checks if the payment method is from a partner integration
func IsPartnerPaymentMethod(method PaymentMethod) bool {
	switch method {
	case PaymentMethodShopify, PaymentMethodPartner:
		return true
	default:
		return false
	}
}

// IsEcommercePlatformPayment checks if the payment method is from an e-commerce platform
func IsEcommercePlatformPayment(method PaymentMethod) bool {
	return GetPaymentMethodCategory(method) == "ECOMMERCE_PLATFORM"
}

// IsMobilePaymentMethod checks if the payment method is a mobile payment solution
func IsMobilePaymentMethod(method PaymentMethod) bool {
	return GetPaymentMethodCategory(method) == "MOBILE_PAYMENT"
}

// IsBuyNowPayLater checks if the payment method is a buy-now-pay-later service
func IsBuyNowPayLater(method PaymentMethod) bool {
	return GetPaymentMethodCategory(method) == "BUY_NOW_PAY_LATER"
}

// GetSupportedPaymentMethods returns a list of all supported payment methods
func GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{
		PaymentMethodCreditCard,
		PaymentMethodPayPal,
		PaymentMethodStripe,
		PaymentMethodBankTransfer,
		PaymentMethodFree,
		PaymentMethodCrypto,
		PaymentMethodShopify,
		PaymentMethodPartner,
	}
}
