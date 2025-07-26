package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	m "github.com/mbbipul/tryonmagic-backend/src/middleware"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/service"
)

func AIModelRoutes(v1 fiber.Router, a service.AuthService, u service.UserService, ai service.AIModelService, s service.SubscriptionService) {
	aiModelController := controller.NewAIModelController(ai)

	aiModels := v1.Group("/ai-models")
	aiModels.Use(m.Auth(a, u))
	aiModels.Post("/", m.PointCheckMiddleware(s, model.TaskTypeGenerateAIModel), aiModelController.CreateAIModel)
	aiModels.Post("/bulk", aiModelController.BulkCreateAIModels)
	aiModels.Get("/", aiModelController.GetAIModels)
	aiModels.Get("/:id", aiModelController.GetAIModelByID)
	// aiModels.Patch("/:id", aiModelController.UpdateAIModel)
	aiModels.Delete("/:id", aiModelController.DeleteAIModel)
}
