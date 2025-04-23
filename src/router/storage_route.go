package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/prodigyai-backend/src/config"
	"github.com/mbbipul/prodigyai-backend/src/controller"
	m "github.com/mbbipul/prodigyai-backend/src/middleware"
	"github.com/mbbipul/prodigyai-backend/src/service"
)

func StorageRoutes(
	v1 fiber.Router, a service.AuthService, u service.UserService, s service.StorageService,
) {
	storageController := controller.NewStorageController(s)
	config.GoogleConfig()

	auth := v1.Group("/storage")

	auth.Post("/save-file", m.Auth(a, u), storageController.SaveFile)
	auth.Get("/files", m.Auth(a, u), storageController.GetFiles)
}
