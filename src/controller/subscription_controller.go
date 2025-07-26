package controller

import (
	"strconv"

	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/service"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SubscriptionController struct {
	SubscriptionService service.SubscriptionService
}

func NewSubscriptionController(subscriptionService service.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{
		SubscriptionService: subscriptionService,
	}
}

// @Summary Get user subscription status
// @Description Get current subscription status and point balance
// @Accept json
// @Produce json
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/status [get]
func (s *SubscriptionController) GetSubscriptionStatus(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	subscription, err := s.SubscriptionService.GetSubscriptionStatus(user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve subscription status")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Subscription status retrieved successfully",
		Code:    fiber.StatusOK,
		Data:    subscription,
	})
}

// @Summary Update subscription plan
// @Description Update user's subscription plan
// @Accept json
// @Produce json
// @Param plan body string true "Subscription plan (FREE, BASIC, PRO, ENTERPRISE)"
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 400 {object} response.GenericErrorResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/plan [put]
func (s *SubscriptionController) UpdateSubscriptionPlan(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	var request struct {
		Plan string `json:"plan"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	plan := model.SubscriptionPlan(request.Plan)
	if !isValidPlan(plan) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid subscription plan")
	}

	subscription, err := s.SubscriptionService.CreateOrUpdateSubscription(user.ID, plan)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update subscription plan")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Subscription plan updated successfully",
		Code:    fiber.StatusOK,
		Data:    subscription,
	})
}

// @Summary Get transaction history
// @Description Get user's comprehensive transaction history
// @Accept json
// @Produce json
// @Param limit query int false "Number of transactions to return (default: 50)"
// @Param offset query int false "Number of transactions to skip (default: 0)"
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/transactions [get]
func (s *SubscriptionController) GetTransactionHistory(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Cap at 100 transactions
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	transactions, err := s.SubscriptionService.GetTransactionHistory(user.ID, limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve transaction history")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Transaction history retrieved successfully",
		Code:    fiber.StatusOK,
		Data:    transactions,
	})
}

// @Summary Get transaction history by type
// @Description Get user's transaction history filtered by transaction type
// @Accept json
// @Produce json
// @Param type query string true "Transaction type (TASK_EXECUTION, SUBSCRIPTION_PURCHASE, POINT_PURCHASE, REFUND, BONUS, etc.)"
// @Param limit query int false "Number of transactions to return (default: 50)"
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 400 {object} response.GenericErrorResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/transactions/type [get]
func (s *SubscriptionController) GetTransactionHistoryByType(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	transactionTypeStr := c.Query("type")
	if transactionTypeStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transaction type is required")
	}

	// Convert string to enum type
	transactionType := model.TransactionType(transactionTypeStr)
	if !model.IsValidTransactionType(transactionType) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid transaction type")
	}

	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Cap at 100 transactions
	}

	transactions, err := s.SubscriptionService.GetTransactionHistoryByType(user.ID, transactionType, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve transaction history")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Transaction history retrieved successfully",
		Code:    fiber.StatusOK,
		Data:    transactions,
	})
}

// @Summary Get transaction summary
// @Description Get user's transaction summary and statistics
// @Accept json
// @Produce json
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/transactions/summary [get]
func (s *SubscriptionController) GetTransactionSummary(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	summary, err := s.SubscriptionService.GetTransactionSummary(user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve transaction summary")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Transaction summary retrieved successfully",
		Code:    fiber.StatusOK,
		Data:    summary,
	})
}

// @Summary Purchase points
// @Description Purchase additional points
// @Accept json
// @Produce json
// @Param points body int true "Number of points to purchase"
// @Param paymentMethod body string true "Payment method (CREDIT_CARD, PAYPAL, STRIPE, SHOPIFY, WOOCOMMERCE, etc.)"
// @Param amountPaid body float64 true "Amount paid in USD"
// @Param referenceId body string true "Payment reference ID"
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 400 {object} response.GenericErrorResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/purchase-points [post]
func (s *SubscriptionController) PurchasePoints(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	var request struct {
		Points        int     `json:"points"`
		PaymentMethod string  `json:"paymentMethod"`
		AmountPaid    float64 `json:"amountPaid"`
		ReferenceID   string  `json:"referenceId"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if request.Points <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Points must be positive")
	}

	if request.AmountPaid <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Amount paid must be positive")
	}

	if request.ReferenceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Reference ID is required")
	}

	// Convert string to enum type
	paymentMethod := model.PaymentMethod(request.PaymentMethod)
	if !model.IsValidPaymentMethod(paymentMethod) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid payment method")
	}

	err := s.SubscriptionService.PurchasePoints(user.ID, request.Points, paymentMethod, request.AmountPaid, request.ReferenceID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to purchase points")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Points purchased successfully",
		Code:    fiber.StatusOK,
		Data:    map[string]interface{}{"pointsPurchased": request.Points, "amountPaid": request.AmountPaid},
	})
}

// @Summary Purchase subscription
// @Description Purchase a subscription plan
// @Accept json
// @Produce json
// @Param plan body string true "Subscription plan (FREE, BASIC, PRO, ENTERPRISE)"
// @Param paymentMethod body string true "Payment method (CREDIT_CARD, PAYPAL, STRIPE, SHOPIFY, WOOCOMMERCE, etc.)"
// @Param referenceId body string true "Payment reference ID"
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 400 {object} response.GenericErrorResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/purchase [post]
func (s *SubscriptionController) PurchaseSubscription(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	var request validation.SubscriptionPurchaseRequest

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err := s.SubscriptionService.PurchaseSubscription(user.ID, request)
	if err != nil {
		if fiberErr, ok := err.(*fiber.Error); ok {
			return fiberErr
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to purchase subscription")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Subscription purchased successfully",
		Code:    fiber.StatusOK,
		Data:    map[string]interface{}{"plan": request.Plan, "amountPaid": model.GetPlanPrice(request.Plan)},
	})
}

// @Summary Add points (Admin only)
// @Description Add points to user's account (Admin only)
// @Accept json
// @Produce json
// @Param points body int true "Number of points to add"
// @Param description body string true "Description of the transaction"
// @Param transactionType body string true "Transaction type (BONUS, REFERRAL_BONUS, PROMOTIONAL_CREDIT, COMPENSATION, ADJUSTMENT)"
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 400 {object} response.GenericErrorResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 403 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/admin/add-points [post]
func (s *SubscriptionController) AddPoints(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// Check if user is admin
	if user.Role != model.RoleAdmin {
		return fiber.NewError(fiber.StatusForbidden, "Admin access required")
	}

	var request struct {
		UserID          string `json:"userId"`
		Points          int    `json:"points"`
		Description     string `json:"description"`
		TransactionType string `json:"transactionType"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if request.Points <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Points must be positive")
	}

	if request.Description == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Description is required")
	}

	// Validate transaction type
	transactionType := model.TransactionType(request.TransactionType)
	validTypes := []model.TransactionType{
		model.TransactionTypeBonus,
		model.TransactionTypeReferralBonus,
		model.TransactionTypePromotionalCredit,
		model.TransactionTypeCompensation,
		model.TransactionTypeAdjustment,
	}

	isValid := false
	for _, validType := range validTypes {
		if transactionType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid transaction type")
	}

	// Parse user ID
	userID, err := parseUUID(request.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	err = s.SubscriptionService.AddPoints(userID, request.Points, request.Description, transactionType)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to add points")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Points added successfully",
		Code:    fiber.StatusOK,
		Data:    map[string]interface{}{"pointsAdded": request.Points},
	})
}

// @Summary Refund points
// @Description Refund points for a failed task
// @Accept json
// @Produce json
// @Param taskId body string true "Task ID to refund"
// @Param points body int true "Number of points to refund"
// @Param reason body string true "Reason for refund"
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 400 {object} response.GenericErrorResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /subscription/refund [post]
func (s *SubscriptionController) RefundPoints(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	var request struct {
		TaskID string `json:"taskId"`
		Points int    `json:"points"`
		Reason string `json:"reason"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if request.Points <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Points must be positive")
	}

	if request.Reason == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Reason is required")
	}

	// Parse task ID
	taskID, err := parseUUID(request.TaskID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	err = s.SubscriptionService.RefundPoints(user.ID, taskID, request.Points, request.Reason)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to refund points")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Points refunded successfully",
		Code:    fiber.StatusOK,
		Data:    map[string]interface{}{"pointsRefunded": request.Points},
	})
}

// Helper functions
func isValidPlan(plan model.SubscriptionPlan) bool {
	switch plan {
	case model.SubscriptionPlanFree, model.SubscriptionPlanBasic, model.SubscriptionPlanPro, model.SubscriptionPlanEnterprise:
		return true
	default:
		return false
	}
}

func parseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}
