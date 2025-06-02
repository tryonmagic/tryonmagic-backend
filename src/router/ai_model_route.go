package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	m "github.com/mbbipul/tryonmagic-backend/src/middleware"
	"github.com/mbbipul/tryonmagic-backend/src/service"
)

func AIModelRoutes(v1 fiber.Router, a service.AuthService, u service.UserService, ai service.AIModelService) {
	aiModelController := controller.NewAIModelController(ai)

	auth := v1.Group("/ai-models")
	auth.Post("/", m.Auth(a, u), aiModelController.CreateAIModel)
	auth.Post("/bulk", m.Auth(a, u), aiModelController.BulkCreateAIModels)
	auth.Get("/", m.Auth(a, u), aiModelController.GetAIModels)
	auth.Get("/:id", m.Auth(a, u), aiModelController.GetAIModelByID)
	// auth.Patch("/:id", m.Auth(a, u), aiModelController.UpdateAIModel)
	auth.Delete("/:id", m.Auth(a, u), aiModelController.DeleteAIModel)
}
