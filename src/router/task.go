package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/config"
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	m "github.com/mbbipul/tryonmagic-backend/src/middleware"
	"github.com/mbbipul/tryonmagic-backend/src/service"
)

func TaskRoutes(
	v1 fiber.Router, a service.AuthService, u service.UserService, t service.TaskService, s service.SubscriptionService,
) {
	taskController := controller.NewTaskController(t)
	config.GoogleConfig()

	tasks := v1.Group("/tasks")
	// Task creation with point checking
	tasks.Use(m.Auth(a, u))
	// taskCreation.Use(m.AdminBypassMiddleware()) // Allow admin bypass
	tasks.Use(m.PointCheckMiddleware(s, nil))
	tasks.Post("/new", taskController.NewTask)
}
