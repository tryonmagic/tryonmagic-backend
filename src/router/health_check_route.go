package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/prodigyai-backend/src/controller"
	"github.com/mbbipul/prodigyai-backend/src/service"
)

func HealthCheckRoutes(v1 fiber.Router, h service.HealthCheckService) {
	healthCheckController := controller.NewHealthCheckController(h)

	healthCheck := v1.Group("/health-check")
	healthCheck.Get("/", healthCheckController.Check)
}
