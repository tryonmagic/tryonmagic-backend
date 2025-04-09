package router

import (
	"app/src/controller"
	m "app/src/middleware"
	"app/src/service"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(v1 fiber.Router, a service.AuthService, u service.UserService, t service.TokenService) {
	userController := controller.NewUserController(u, t)

	user := v1.Group("/users")

	user.Get("/", m.Auth(a, u, "getUsers"), userController.GetUsers)
	user.Post("/", m.Auth(a, u, "manageUsers"), userController.CreateUser)
	user.Get("/:userId", m.Auth(a, u, "getUsers"), userController.GetUserByID)
	user.Patch("/:userId", m.Auth(a, u, "manageUsers"), userController.UpdateUser)
	user.Delete("/:userId", m.Auth(a, u, "manageUsers"), userController.DeleteUser)
}
