package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SubscriptionService interface {
	GetUserSubscription(userID uuid.UUID) (*model.Subscription, error)
	CreateOrUpdateSubscription(userID uuid.UUID, plan model.SubscriptionPlan) (*model.Subscription, error)
	DeductPoints(userID uuid.UUID, taskID uuid.UUID, points int, taskType model.TaskType) error
	AddPoints(userID uuid.UUID, points int, description string, transactionType model.TransactionType) error
	RefundPoints(userID uuid.UUID, taskID uuid.UUID, points int, reason string) error
	CheckUserCanPerformTask(userID uuid.UUID, requiredPoints int) (bool, error)
	ResetMonthlyPoints(userID uuid.UUID) error
	GetTransactionHistory(userID uuid.UUID, limit int, offset int) ([]model.TransactionHistory, error)
	GetTransactionHistoryByType(userID uuid.UUID, transactionType model.TransactionType, limit int) ([]model.TransactionHistory, error)
	PurchasePoints(userID uuid.UUID, points int, paymentMethod model.PaymentMethod, amountPaid float64, referenceID string) error
	PurchaseSubscription(userID uuid.UUID, paymentRequest validation.SubscriptionPurchaseRequest) error
	GetSubscriptionStatus(userID uuid.UUID) (*model.Subscription, error)
	GetTransactionSummary(userID uuid.UUID) (map[string]interface{}, error)
}

type subscriptionService struct {
	Log *logrus.Logger
	DB  *gorm.DB
}

func NewSubscriptionService(db *gorm.DB) SubscriptionService {
	return &subscriptionService{
		Log: utils.Log,
		DB:  db,
	}
}

func (s *subscriptionService) GetUserSubscription(userID uuid.UUID) (*model.Subscription, error) {
	var subscription model.Subscription
	err := s.DB.WithContext(context.Background()).
		Where("user_id = ?", userID).
		First(&subscription).Error

	if err != nil {
		// ToDo:Business Decision
		// if errors.Is(err, gorm.ErrRecordNotFound) {
		// 	// Create default free subscription if none exists
		// 	return s.CreateOrUpdateSubscription(userID, model.SubscriptionPlanFree)
		// }
		return nil, fiber.NewError(fiber.StatusForbidden, "Please purchase a subscription plan")
	}

	// ToDo: Business Decision (one time free subscription?)
	// Check if monthly reset is needed
	if time.Now().After(subscription.NextResetDate) {
		if err := s.ResetMonthlyPoints(userID); err != nil {
			s.Log.Errorf("Failed to reset monthly points: %v", err)
		}
		// Refresh subscription data
		return s.GetUserSubscription(userID)
	}

	return &subscription, nil
}

func (s *subscriptionService) CreateOrUpdateSubscription(userID uuid.UUID, plan model.SubscriptionPlan) (*model.Subscription, error) {
	var subscription model.Subscription

	err := s.DB.WithContext(context.Background()).
		Where("user_id = ?", userID).
		First(&subscription).Error

	monthlyPoints := model.GetSubscriptionPlanPoints(plan)
	now := time.Now()
	nextReset := now.AddDate(0, 1, 0)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new subscription
			subscription = model.Subscription{
				UserID:        userID,
				Plan:          plan,
				Status:        model.SubscriptionStatusActive,
				PointsBalance: monthlyPoints,
				PointsUsed:    0,
				PointsTotal:   monthlyPoints,
				MonthlyPoints: monthlyPoints,
				LastResetDate: now,
				NextResetDate: nextReset,
			}

			if err := s.DB.WithContext(context.Background()).Create(&subscription).Error; err != nil {
				return nil, err
			}

			// Log initial points allocation
			s.logTransactionHistory(userID, nil, model.TransactionTypeSubscriptionPurchase,
				monthlyPoints, 0, fmt.Sprintf("Initial subscription: %s plan", plan),
				"", model.ReferenceTypeSubscription, model.PaymentMethodFree,
				model.PaymentStatusCompleted, 0.0, "USD", 1.0, nil)
		} else {
			return nil, err
		}
	} else {
		// Update existing subscription
		oldPlan := subscription.Plan
		subscription.Plan = plan
		subscription.MonthlyPoints = monthlyPoints
		subscription.NextResetDate = nextReset

		if err := s.DB.WithContext(context.Background()).Save(&subscription).Error; err != nil {
			return nil, err
		}

		// Log plan change if different
		if oldPlan != plan {
			transactionType := model.TransactionTypePlanUpgrade
			if model.GetSubscriptionPlanPoints(oldPlan) > model.GetSubscriptionPlanPoints(plan) {
				transactionType = model.TransactionTypePlanDowngrade
			}
			s.logTransactionHistory(userID, nil, transactionType,
				monthlyPoints, 0, fmt.Sprintf("Plan change from %s to %s", oldPlan, plan),
				"", model.ReferenceTypeSubscription, model.PaymentMethodFree,
				model.PaymentStatusCompleted, 0.0, "USD", 1.0, nil)
		}
	}

	return &subscription, nil
}

func (s *subscriptionService) DeductPoints(userID uuid.UUID, taskID uuid.UUID, points int, taskType model.TaskType) error {
	return s.DB.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var subscription model.Subscription
		if err := tx.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
			return err
		}

		if !subscription.CanPerformTask(points) {
			return errors.New("insufficient points or inactive subscription")
		}

		subscription.PointsBalance -= points
		subscription.PointsUsed += points

		if err := tx.Save(&subscription).Error; err != nil {
			return err
		}

		// Log point deduction transaction
		description := fmt.Sprintf("Task execution: %s", taskType)
		return s.logTransactionHistoryWithTx(tx, userID, &taskID, model.TransactionTypeTaskExecution,
			points, 0, description, taskID.String(), model.ReferenceTypeTask,
			model.PaymentMethodFree, model.PaymentStatusCompleted, 0.0, "USD", 1.0, nil)
	})
}

func (s *subscriptionService) AddPoints(userID uuid.UUID, points int, description string, transactionType model.TransactionType) error {
	return s.DB.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var subscription model.Subscription
		if err := tx.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
			return err
		}

		subscription.PointsBalance += points
		subscription.PointsTotal += points

		if err := tx.Save(&subscription).Error; err != nil {
			return err
		}

		return s.logTransactionHistoryWithTx(tx, userID, nil, transactionType,
			0, points, description, "", model.ReferenceTypeBonus,
			model.PaymentMethodFree, model.PaymentStatusCompleted, 0.0, "USD", 1.0, nil)
	})
}

func (s *subscriptionService) RefundPoints(userID uuid.UUID, taskID uuid.UUID, points int, reason string) error {
	return s.DB.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var subscription model.Subscription
		if err := tx.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
			return err
		}

		subscription.PointsBalance += points
		subscription.PointsUsed -= points

		if err := tx.Save(&subscription).Error; err != nil {
			return err
		}

		description := fmt.Sprintf("Refund: %s", reason)
		return s.logTransactionHistoryWithTx(tx, userID, &taskID, model.TransactionTypeRefund,
			0, points, description, taskID.String(), model.ReferenceTypeRefund,
			model.PaymentMethodFree, model.PaymentStatusCompleted, 0.0, "USD", 1.0, nil)
	})
}

func (s *subscriptionService) PurchasePoints(userID uuid.UUID, points int, paymentMethod model.PaymentMethod, amountPaid float64, referenceID string) error {
	return s.DB.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var subscription model.Subscription
		if err := tx.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
			return err
		}

		subscription.PointsBalance += points
		subscription.PointsTotal += points

		if err := tx.Save(&subscription).Error; err != nil {
			return err
		}

		description := fmt.Sprintf("Purchased %d points", points)
		exchangeRate := amountPaid / float64(points)

		return s.logTransactionHistoryWithTx(tx, userID, nil, model.TransactionTypePointPurchase,
			0, points, description, referenceID, model.ReferenceTypePayment,
			paymentMethod, model.PaymentStatusCompleted, amountPaid, "USD", exchangeRate, nil)
	})
}

func (s *subscriptionService) PurchaseSubscription(userID uuid.UUID, paymentRequest validation.SubscriptionPurchaseRequest) error {
	return s.DB.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var subscription model.Subscription
		// Check if there are already 100 free subscriptions before allowing another
		if paymentRequest.Plan == model.SubscriptionPlanFree {
			var freeSubsCount int64
			if err := tx.Model(&model.Subscription{}).
				Where("plan = ?", model.SubscriptionPlanFree).
				Count(&freeSubsCount).Error; err != nil {
				return err
			}
			if freeSubsCount >= 100 {
				return fiber.NewError(fiber.StatusForbidden, "The free limited time offer is no longer available. Currently, we gave the first 100 users a limited credit of 20 points to try our app. Soon, paid subscription plans will be published for more credit.")
			}
		}

		if err := tx.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// User doesn't have a subscription, allow purchase
				subscription = model.Subscription{
					UserID:        userID,
					Plan:          paymentRequest.Plan,
					Status:        model.SubscriptionStatusActive,
					PointsUsed:    0,
					LastResetDate: time.Now(),
				}
				if err := tx.Create(&subscription).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// User already has a subscription, check if it's active for current month
			if
			// subscription.Plan == paymentRequest.Plan && // ToDo: Business Decision
			subscription.Status == model.SubscriptionStatusActive &&
				subscription.NextResetDate.After(time.Now()) {
				return fiber.NewError(fiber.StatusBadRequest, "You already have an active subscription for this month. Please purchase points.")
			}
		}

		monthlyPoints := model.GetSubscriptionPlanPoints(paymentRequest.Plan)
		amountPaid := model.GetPlanPrice(paymentRequest.Plan)

		subscription.Plan = paymentRequest.Plan
		subscription.MonthlyPoints = monthlyPoints
		// Instead of replacing, add the new monthlyPoints to the existing balance
		subscription.PointsBalance += monthlyPoints
		subscription.PointsTotal += monthlyPoints
		subscription.NextResetDate = time.Now().AddDate(0, 1, 0)
		subscription.ExpiresAt = &subscription.NextResetDate

		if err := tx.Save(&subscription).Error; err != nil {
			return err
		}

		description := fmt.Sprintf("Subscription purchase: %s plan", paymentRequest.Plan)
		exchangeRate := amountPaid / float64(monthlyPoints)

		return s.logTransactionHistoryWithTx(tx, userID, nil, model.TransactionTypeSubscriptionPurchase,
			0, monthlyPoints, description, paymentRequest.ReferenceID, model.ReferenceTypeSubscription,
			paymentRequest.PaymentMethod, model.PaymentStatusCompleted, amountPaid, "USD", exchangeRate, nil)
	})
}

func (s *subscriptionService) CheckUserCanPerformTask(userID uuid.UUID, requiredPoints int) (bool, error) {
	subscription, err := s.GetUserSubscription(userID)
	if err != nil {
		return false, err
	}

	return subscription.CanPerformTask(requiredPoints), nil
}

func (s *subscriptionService) ResetMonthlyPoints(userID uuid.UUID) error {
	return s.DB.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var subscription model.Subscription
		if err := tx.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
			return err
		}

		subscription.PointsBalance = subscription.MonthlyPoints
		subscription.PointsUsed = 0
		subscription.LastResetDate = time.Now()
		subscription.NextResetDate = time.Now().AddDate(0, 1, 0)

		if err := tx.Save(&subscription).Error; err != nil {
			return err
		}

		// Log monthly reset
		description := fmt.Sprintf("Monthly points reset for %s plan", subscription.Plan)
		return s.logTransactionHistoryWithTx(tx, userID, nil, model.TransactionTypeMonthlyReset,
			0, subscription.MonthlyPoints, description, "", model.ReferenceTypeSubscription,
			model.PaymentMethodFree, model.PaymentStatusCompleted, 0.0, "USD", 1.0, nil)
	})
}

func (s *subscriptionService) GetTransactionHistory(userID uuid.UUID, limit int, offset int) ([]model.TransactionHistory, error) {
	var transactions []model.TransactionHistory
	err := s.DB.WithContext(context.Background()).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, err
}

func (s *subscriptionService) GetTransactionHistoryByType(userID uuid.UUID, transactionType model.TransactionType, limit int) ([]model.TransactionHistory, error) {
	var transactions []model.TransactionHistory
	err := s.DB.WithContext(context.Background()).
		Where("user_id = ? AND transaction_type = ?", userID, transactionType).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions).Error

	return transactions, err
}

func (s *subscriptionService) GetSubscriptionStatus(userID uuid.UUID) (*model.Subscription, error) {
	return s.GetUserSubscription(userID)
}

func (s *subscriptionService) GetTransactionSummary(userID uuid.UUID) (map[string]interface{}, error) {
	var result struct {
		TotalTransactions int64   `json:"totalTransactions"`
		TotalDebits       int64   `json:"totalDebits"`
		TotalCredits      int64   `json:"totalCredits"`
		TotalAmountPaid   float64 `json:"totalAmountPaid"`
		TaskExecutions    int64   `json:"taskExecutions"`
		Purchases         int64   `json:"purchases"`
		Refunds           int64   `json:"refunds"`
		Bonuses           int64   `json:"bonuses"`
	}

	err := s.DB.WithContext(context.Background()).
		Model(&model.TransactionHistory{}).
		Select(`
			COUNT(*) as total_transactions,
			SUM(debit_amount) as total_debits,
			SUM(credit_amount) as total_credits,
			SUM(amount_paid) as total_amount_paid,
			SUM(CASE WHEN transaction_type = ? THEN 1 ELSE 0 END) as task_executions,
			SUM(CASE WHEN transaction_type IN (?, ?) THEN 1 ELSE 0 END) as purchases,
			SUM(CASE WHEN transaction_type = ? THEN 1 ELSE 0 END) as refunds,
			SUM(CASE WHEN transaction_type IN (?, ?, ?) THEN 1 ELSE 0 END) as bonuses
		`,
			model.TransactionTypeTaskExecution,
			model.TransactionTypeSubscriptionPurchase, model.TransactionTypePointPurchase,
			model.TransactionTypeRefund,
			model.TransactionTypeBonus, model.TransactionTypeReferralBonus, model.TransactionTypePromotionalCredit).
		Where("user_id = ?", userID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalTransactions": result.TotalTransactions,
		"totalDebits":       result.TotalDebits,
		"totalCredits":      result.TotalCredits,
		"totalAmountPaid":   result.TotalAmountPaid,
		"taskExecutions":    result.TaskExecutions,
		"purchases":         result.Purchases,
		"refunds":           result.Refunds,
		"bonuses":           result.Bonuses,
		"netPoints":         result.TotalCredits - result.TotalDebits,
	}, nil
}

func (s *subscriptionService) logTransactionHistory(userID uuid.UUID, taskID *uuid.UUID, transactionType model.TransactionType,
	debitAmount, creditAmount int, description, referenceID string, referenceType model.ReferenceType, paymentMethod model.PaymentMethod, paymentStatus model.PaymentStatus,
	amountPaid float64, currency string, exchangeRate float64, metadata map[string]interface{}) error {

	var subscription model.Subscription
	if err := s.DB.WithContext(context.Background()).Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		return err
	}

	balanceBefore := subscription.PointsBalance
	balanceAfter := subscription.PointsBalance

	// Adjust balance based on transaction type
	if debitAmount > 0 {
		balanceBefore = subscription.PointsBalance + debitAmount
		balanceAfter = subscription.PointsBalance
	} else if creditAmount > 0 {
		balanceBefore = subscription.PointsBalance - creditAmount
		balanceAfter = subscription.PointsBalance
	}

	metadataJSON := ""
	if metadata != nil {
		if metadataBytes, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(metadataBytes)
		}
	}

	transaction := model.TransactionHistory{
		UserID:            userID,
		TaskID:            taskID,
		TransactionType:   transactionType,
		TransactionStatus: model.TransactionStatusCompleted,
		DebitAmount:       debitAmount,
		CreditAmount:      creditAmount,
		NetAmount:         creditAmount - debitAmount,
		BalanceBefore:     balanceBefore,
		BalanceAfter:      balanceAfter,
		Description:       description,
		ReferenceID:       referenceID,
		ReferenceType:     referenceType,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     paymentStatus,
		Currency:          currency,
		AmountPaid:        amountPaid,
		ExchangeRate:      exchangeRate,
		Metadata:          metadataJSON,
	}

	return s.DB.WithContext(context.Background()).Create(&transaction).Error
}

func (s *subscriptionService) logTransactionHistoryWithTx(tx *gorm.DB, userID uuid.UUID, taskID *uuid.UUID, transactionType model.TransactionType,
	debitAmount, creditAmount int, description, referenceID string, referenceType model.ReferenceType, paymentMethod model.PaymentMethod, paymentStatus model.PaymentStatus,
	amountPaid float64, currency string, exchangeRate float64, metadata map[string]interface{}) error {

	var subscription model.Subscription
	if err := tx.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		return err
	}

	balanceBefore := subscription.PointsBalance
	balanceAfter := subscription.PointsBalance

	// Adjust balance based on transaction type
	if debitAmount > 0 {
		balanceBefore = subscription.PointsBalance + debitAmount
		balanceAfter = subscription.PointsBalance
	} else if creditAmount > 0 {
		balanceBefore = subscription.PointsBalance - creditAmount
		balanceAfter = subscription.PointsBalance
	}

	metadataJSON := ""
	if metadata != nil {
		if metadataBytes, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(metadataBytes)
		}
	}

	transaction := model.TransactionHistory{
		UserID:            userID,
		TaskID:            taskID,
		TransactionType:   transactionType,
		TransactionStatus: model.TransactionStatusCompleted,
		DebitAmount:       debitAmount,
		CreditAmount:      creditAmount,
		NetAmount:         creditAmount - debitAmount,
		BalanceBefore:     balanceBefore,
		BalanceAfter:      balanceAfter,
		Description:       description,
		ReferenceID:       referenceID,
		ReferenceType:     referenceType,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     paymentStatus,
		Currency:          currency,
		AmountPaid:        amountPaid,
		ExchangeRate:      exchangeRate,
		Metadata:          metadataJSON,
	}

	return tx.Create(&transaction).Error
}
