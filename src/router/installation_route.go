package router

import (
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	"github.com/mbbipul/tryonmagic-backend/src/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// InstallationRoutes sets up app installation-related API routes
func InstallationRoutes(v1 fiber.Router, installationService service.AppInstallationService, validate *validator.Validate) {
	installationController := controller.NewAppInstallationController(installationService, validate)

	// App Installation API routes - Base URL: /api/installations
	installations := v1.Group("/installations")

	// Installation Management
	installations.Get("/:shopDomain", installationController.CheckInstallation)
	installations.Post("/", installationController.CreateInstallation)
	installations.Patch("/:shopDomain", installationController.UpdateInstallation)
	installations.Delete("/:shopDomain", installationController.DeleteInstallation)
	installations.Get("/:shopDomain/details", installationController.GetInstallation)
}
