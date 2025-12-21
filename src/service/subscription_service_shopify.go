package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
	"gorm.io/gorm"
)

// SyncShopifySubscription syncs subscription data from Shopify
func (s *subscriptionService) SyncShopifySubscription(userID uuid.UUID, syncRequest validation.ShopifySyncRequest) (*model.Subscription, error) {
	var subscription model.Subscription

	// Convert string values to enum types
	plan := model.SubscriptionPlan(syncRequest.Plan)
	status := model.SubscriptionStatus(syncRequest.Status)

	// Validate plan and status
	if !model.IsValidSubscriptionPlan(plan) {
		return nil, errors.New("invalid subscription plan")
	}
	if !model.IsValidSubscriptionStatus(status) {
		return nil, errors.New("invalid subscription status")
	}

	err := s.DB.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		// Check if subscription exists
		err := tx.Where("user_id = ?", userID).First(&subscription).Error

		monthlyPoints := model.GetSubscriptionPlanPoints(plan)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new subscription
				subscription = model.Subscription{
					UserID:        userID,
					Plan:          plan,
					Status:        status,
					PointsBalance: monthlyPoints,
					PointsUsed:    0,
					PointsTotal:   monthlyPoints,
					MonthlyPoints: monthlyPoints,
					LastResetDate: syncRequest.CurrentPeriodStart,
					NextResetDate: syncRequest.CurrentPeriodEnd,
					ExpiresAt:     &syncRequest.CurrentPeriodEnd,
				}

				if err := tx.Create(&subscription).Error; err != nil {
					return err
				}

				// Log Shopify sync transaction
				metadata := map[string]interface{}{
					"shopDomain": syncRequest.ShopDomain,
					"source":     "shopify_sync",
				}
				if syncRequest.Metadata != nil {
					for k, v := range syncRequest.Metadata {
						metadata[k] = v
					}
				}

				return s.logTransactionHistoryWithTx(tx, userID, nil, model.TransactionTypeSubscriptionPurchase,
					0, monthlyPoints, fmt.Sprintf("Shopify subscription sync: %s plan", plan),
					syncRequest.ShopDomain, model.ReferenceTypeSubscription, model.PaymentMethodShopify,
					model.PaymentStatusCompleted, 0.0, "USD", 1.0, metadata)
			}
			return err
		}

		// Update existing subscription
		oldPlan := subscription.Plan
		oldStatus := subscription.Status

		subscription.Plan = plan
		subscription.Status = status
		subscription.MonthlyPoints = monthlyPoints
		subscription.LastResetDate = syncRequest.CurrentPeriodStart
		subscription.NextResetDate = syncRequest.CurrentPeriodEnd
		subscription.ExpiresAt = &syncRequest.CurrentPeriodEnd

		// If plan changed and status is active, add new points
		if oldPlan != plan && status == model.SubscriptionStatusActive {
			pointsDiff := monthlyPoints - model.GetSubscriptionPlanPoints(oldPlan)
			if pointsDiff > 0 {
				subscription.PointsBalance += pointsDiff
				subscription.PointsTotal += pointsDiff
			}
		}

		if err := tx.Save(&subscription).Error; err != nil {
			return err
		}

		// Log sync transaction
		metadata := map[string]interface{}{
			"shopDomain": syncRequest.ShopDomain,
			"source":     "shopify_sync",
			"oldPlan":    oldPlan,
			"newPlan":    plan,
			"oldStatus":  oldStatus,
			"newStatus":  status,
		}
		if syncRequest.Metadata != nil {
			for k, v := range syncRequest.Metadata {
				metadata[k] = v
			}
		}

		transactionType := model.TransactionTypeSubscriptionPurchase
		if oldPlan != plan {
			if model.GetSubscriptionPlanPoints(plan) > model.GetSubscriptionPlanPoints(oldPlan) {
				transactionType = model.TransactionTypePlanUpgrade
			} else {
				transactionType = model.TransactionTypePlanDowngrade
			}
		}

		description := fmt.Sprintf("Shopify subscription sync: %s plan (status: %s)", plan, status)
		return s.logTransactionHistoryWithTx(tx, userID, nil, transactionType,
			0, 0, description, syncRequest.ShopDomain, model.ReferenceTypeSubscription,
			model.PaymentMethodShopify, model.PaymentStatusCompleted, 0.0, "USD", 1.0, metadata)
	})

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}
