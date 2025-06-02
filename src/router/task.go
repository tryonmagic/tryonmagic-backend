package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/config"
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	m "github.com/mbbipul/tryonmagic-backend/src/middleware"
	"github.com/mbbipul/tryonmagic-backend/src/service"
)

func TaskRoutes(
	v1 fiber.Router, a service.AuthService, u service.UserService, t service.TaskService,
) {
	taskController := controller.NewTaskController(t)
	config.GoogleConfig()

	auth := v1.Group("/tasks")

	auth.Post("/new", m.Auth(a, u), taskController.NewTask)
}
