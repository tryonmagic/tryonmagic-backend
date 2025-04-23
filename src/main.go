package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mbbipul/prodigyai-backend/src/config"
	"github.com/mbbipul/prodigyai-backend/src/database"
	"github.com/mbbipul/prodigyai-backend/src/internal/adapter/grpc"
	"github.com/mbbipul/prodigyai-backend/src/middleware"
	"github.com/mbbipul/prodigyai-backend/src/router"
	"github.com/mbbipul/prodigyai-backend/src/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"gorm.io/gorm"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := setupFiberApp()
	db := setupDatabase()
	defer closeDatabase(db)
	setupRoutes(app, db)

	address := fmt.Sprintf("%s:%d", config.AppHost, config.AppPort)

	// Start Fiber server and gRPC server
	serverErrors := make(chan error, 2)
	go startServer(app, address, serverErrors)
	go startGrpcServer(db, serverErrors)

	handleGracefulShutdown(ctx, app, serverErrors)
}

func setupFiberApp() *fiber.App {
	app := fiber.New(config.FiberConfig())

	// Middleware setup
	app.Use("/v1/auth", middleware.LimiterConfig())
	app.Use(middleware.LoggerConfig())
	app.Use(helmet.New())
	app.Use(compress.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, x-shopify-session-token",
	}))
	app.Use(middleware.RecoverConfig())

	return app
}

func setupDatabase() *gorm.DB {
	db := database.Connect(config.DBHost, config.DBName)
	// Add any additional database setup if needed
	return db
}

func setupRoutes(app *fiber.App, db *gorm.DB) {
	router.Routes(app, db)
	app.Use(utils.NotFoundHandler)
}

func startServer(app *fiber.App, address string, errs chan<- error) {
	if err := app.Listen(address); err != nil {
		errs <- fmt.Errorf("error starting Fiber server: %w", err)
	}
}

func startGrpcServer(db *gorm.DB, errs chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			errs <- fmt.Errorf("gRPC server panic: %v", r)
		}
	}()
	grpc.StartServer(db)
}

func closeDatabase(db *gorm.DB) {
	sqlDB, errDB := db.DB()
	if errDB != nil {
		utils.Log.Errorf("Error getting database instance: %v", errDB)
		return
	}

	if err := sqlDB.Close(); err != nil {
		utils.Log.Errorf("Error closing database connection: %v", err)
	} else {
		utils.Log.Info("Database connection closed successfully")
	}
}

func handleGracefulShutdown(ctx context.Context, app *fiber.App, serverErrors <-chan error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		utils.Log.Fatalf("Server error: %v", err)
	case <-quit:
		utils.Log.Info("Shutting down servers...")
		if err := app.Shutdown(); err != nil {
			utils.Log.Fatalf("Error during Fiber server shutdown: %v", err)
		}
		// Add any additional cleanup for gRPC server if needed
	case <-ctx.Done():
		utils.Log.Info("Servers exiting due to context cancellation")
	}

	utils.Log.Info("Servers exited")
}
