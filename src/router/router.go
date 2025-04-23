package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/prodigyai-backend/src/config"
	"github.com/mbbipul/prodigyai-backend/src/service"
	"github.com/mbbipul/prodigyai-backend/src/validation"
	"gorm.io/gorm"
)

func Routes(app *fiber.App, db *gorm.DB) {
	validate := validation.Validator()

	healthCheckService := service.NewHealthCheckService(db)
	emailService := service.NewEmailService()
	userService := service.NewUserService(db, validate)
	tokenService := service.NewTokenService(db, validate, userService)
	authService := service.NewAuthService(db, validate, userService, tokenService)
	storageService := service.NewStorageService(db, validate)
	taskService := service.NewTaskService(db, validate)

	v1 := app.Group("/v1")

	HealthCheckRoutes(v1, healthCheckService)
	AuthRoutes(v1, authService, userService, tokenService, emailService)
	UserRoutes(v1, authService, userService, tokenService)
	StorageRoutes(v1, authService, userService, storageService)
	TaskRoutes(v1, authService, userService, taskService)
	// TODO: add another routes here...

	if !config.IsProd {
		DocsRoutes(v1)
	}
}
