package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	m "github.com/mbbipul/tryonmagic-backend/src/middleware"
	"github.com/mbbipul/tryonmagic-backend/src/service"
)

func GarmentRoutes(v1 fiber.Router, a service.AuthService, u service.UserService, garmentService service.GarmentService) {
	garmentController := controller.NewGarmentController(garmentService)
	garment := v1.Group("/garments")

	garment.Get("/", m.Auth(a, u), garmentController.GetGarments)
	garment.Post("/", m.Auth(a, u), garmentController.CreateGarment)
	garment.Get("/:id", m.Auth(a, u), garmentController.GetGarmentByID)
	garment.Patch("/:id", m.Auth(a, u), garmentController.UpdateGarment)
	garment.Delete("/:id", m.Auth(a, u), garmentController.DeleteGarment)
}
