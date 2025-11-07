package router

import (
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	"github.com/mbbipul/tryonmagic-backend/src/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// GDPRRoutes sets up GDPR-related API routes
func GDPRRoutes(v1 fiber.Router, gdprService service.GDPRService, validate *validator.Validate) {
	gdprController := controller.NewGDPRController(gdprService, validate)

	// GDPR API routes - Base URL: /api/gdpr
	gdpr := v1.Group("/gdpr")

	// GDPR Request Management
	gdpr.Post("/requests", gdprController.LogGDPRRequest)
	gdpr.Patch("/requests/:requestId/status", gdprController.UpdateRequestStatus)

	// Customer Data Management
	gdpr.Get("/customers/:customerId", gdprController.GetCustomerData)
	gdpr.Post("/customers/:customerId/redact", gdprController.RedactCustomerData)

	// Shop Data Management
	gdpr.Post("/shops/:shop/redact", gdprController.RedactShopData)

	// Maintenance
	gdpr.Post("/cleanup", gdprController.CleanupOldLogs)
}
