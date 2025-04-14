package router

import (
	"app/src/config"
	"app/src/controller"
	m "app/src/middleware"
	"app/src/service"

	"github.com/gofiber/fiber/v2"
)

func StorageRoutes(
	v1 fiber.Router, a service.AuthService, u service.UserService, s service.StorageService,
) {
	storageController := controller.NewStorageController(s)
	config.GoogleConfig()

	auth := v1.Group("/storage")

	auth.Post("/save-file", m.Auth(a, u), storageController.SaveFile)
}
