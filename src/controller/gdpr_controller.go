package controller

import (
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/service"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// GDPRController handles GDPR-related API endpoints
type GDPRController struct {
	GDPRService service.GDPRService
	Validate    *validator.Validate
}

// NewGDPRController creates a new GDPR controller instance
func NewGDPRController(gdprService service.GDPRService, validate *validator.Validate) *GDPRController {
	return &GDPRController{
		GDPRService: gdprService,
		Validate:    validate,
	}
}

// LogGDPRRequest creates a new GDPR request log entry
// @Summary Log GDPR Request
// @Description Creates a new GDPR request log entry
// @Tags GDPR
// @Accept json
// @Produce json
// @Param request body validation.LogGDPRRequest true "GDPR Request Data"
// @Success 201 {object} validation.GDPRRequestResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /gdpr/requests [post]
func (g *GDPRController) LogGDPRRequest(c *fiber.Ctx) error {
	req := new(validation.LogGDPRRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_BODY",
		})
	}

	if err := g.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "VALIDATION_ERROR",
		})
	}

	gdprResponse, err := g.GDPRService.LogGDPRRequest(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "GDPR request logged successfully",
		Code:    fiber.StatusCreated,
		Data:    gdprResponse,
	})
}

// UpdateRequestStatus updates the status of a GDPR request
// @Summary Update GDPR Request Status
// @Description Updates the status of a GDPR request
// @Tags GDPR
// @Accept json
// @Produce json
// @Param requestId path string true "Request ID"
// @Param request body validation.UpdateGDPRRequestStatus true "Status Update Data"
// @Success 200 {object} validation.GDPRStatusResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /gdpr/requests/{requestId}/status [patch]
func (g *GDPRController) UpdateRequestStatus(c *fiber.Ctx) error {
	requestID := c.Params("requestId")
	if requestID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Request ID is required",
			Code:  "MISSING_REQUEST_ID",
		})
	}

	req := new(validation.UpdateGDPRRequestStatus)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_BODY",
		})
	}

	if err := g.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "VALIDATION_ERROR",
		})
	}

	statusResponse, err := g.GDPRService.UpdateRequestStatus(requestID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "GDPR request status updated successfully",
		Code:    fiber.StatusOK,
		Data:    statusResponse,
	})
}

// GetCustomerData gathers all customer data for a GDPR data request
// @Summary Get Customer Data
// @Description Gathers all customer data for a GDPR data request
// @Tags GDPR
// @Produce json
// @Param customerId path string true "Customer ID"
// @Param shop query string true "Shop Domain"
// @Success 200 {object} validation.CustomerDataResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /gdpr/customers/{customerId} [get]
func (g *GDPRController) GetCustomerData(c *fiber.Ctx) error {
	customerID := c.Params("customerId")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Customer ID is required",
			Code:  "MISSING_CUSTOMER_ID",
		})
	}

	shop := c.Query("shop")
	if shop == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Shop parameter is required",
			Code:  "MISSING_SHOP_PARAM",
		})
	}

	dataResponse, err := g.GDPRService.GetCustomerData(customerID, shop)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Customer data retrieved successfully",
		Code:    fiber.StatusOK,
		Data:    dataResponse,
	})
}

// RedactCustomerData deletes or anonymizes all customer data for GDPR compliance
// @Summary Redact Customer Data
// @Description Deletes or anonymizes all customer data for GDPR compliance
// @Tags GDPR
// @Accept json
// @Produce json
// @Param customerId path string true "Customer ID"
// @Param request body validation.RedactCustomerDataRequest true "Redaction Data"
// @Success 200 {object} validation.RedactionResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /gdpr/customers/{customerId}/redact [post]
func (g *GDPRController) RedactCustomerData(c *fiber.Ctx) error {
	customerID := c.Params("customerId")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Customer ID is required",
			Code:  "MISSING_CUSTOMER_ID",
		})
	}

	req := new(validation.RedactCustomerDataRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_BODY",
		})
	}

	if err := g.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "VALIDATION_ERROR",
		})
	}

	redactResponse, err := g.GDPRService.RedactCustomerData(customerID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Customer data redacted successfully",
		Code:    fiber.StatusOK,
		Data:    redactResponse,
	})
}

// RedactShopData deletes all shop-related data
// @Summary Redact Shop Data
// @Description Deletes all shop-related data. Called 48 hours after app uninstall.
// @Tags GDPR
// @Produce json
// @Param shop path string true "Shop Domain"
// @Success 200 {object} validation.RedactionResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /gdpr/shops/{shop}/redact [post]
func (g *GDPRController) RedactShopData(c *fiber.Ctx) error {
	shop := c.Params("shop")
	if shop == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Shop parameter is required",
			Code:  "MISSING_SHOP_PARAM",
		})
	}

	shopRedactResponse, err := g.GDPRService.RedactShopData(shop)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Shop data redacted successfully",
		Code:    fiber.StatusOK,
		Data:    shopRedactResponse,
	})
}

// CleanupOldLogs removes GDPR request logs older than 2 years
// @Summary Cleanup Old GDPR Logs
// @Description Removes GDPR request logs older than 2 years
// @Tags GDPR
// @Produce json
// @Success 200 {object} validation.CleanupResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /gdpr/cleanup [post]
func (g *GDPRController) CleanupOldLogs(c *fiber.Ctx) error {
	cleanupResponse, err := g.GDPRService.CleanupOldLogs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Old GDPR logs cleaned up successfully",
		Code:    fiber.StatusOK,
		Data:    cleanupResponse,
	})
}
