package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mbbipul/tryonmagic-backend/src/config"
	"github.com/mbbipul/tryonmagic-backend/src/database"
	"github.com/mbbipul/tryonmagic-backend/src/internal/adapter/grpc"
	socket "github.com/mbbipul/tryonmagic-backend/src/internal/adapter/socketio"
	"github.com/mbbipul/tryonmagic-backend/src/middleware"
	"github.com/mbbipul/tryonmagic-backend/src/router"
	"github.com/mbbipul/tryonmagic-backend/src/utils"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
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

	setupSocketIO(app)

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

func setupSocketIO(app *fiber.App) {

	// Middleware to check for WebSocket upgrade
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Multiple event handling supported
	socketio.On(socketio.EventConnect, func(ep *socketio.EventPayload) {
		fmt.Printf("Connection event 1 - User: %s", ep.Kws.GetStringAttribute("user_id"))
	})
	socketio.On(socketio.EventConnect, func(ep *socketio.EventPayload) {
		fmt.Printf("Connection event 2 - User: %s", ep.Kws.GetStringAttribute("user_id"))
	})

	// On message event
	socketio.On(socketio.EventMessage, func(ep *socketio.EventPayload) {
		fmt.Printf("Message event - User: %s - Message: %s", ep.Kws.GetStringAttribute("user_id"), string(ep.Data))

		// var message any

		// // Unmarshal the json message
		// // {
		// //  "from": "<user-id>",
		// //  "to": "<recipient-user-id>",
		// //  "data": "hello"
		// //}
		// err := json.Unmarshal(ep.Data, &message)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// // Emit the message directly to specified user
		// err = ep.Kws.EmitTo(clients[message.To], ep.Data)
		// if err != nil {
		// 	fmt.Println(err)
		// }
	})

	// On disconnect event
	socketio.On(socketio.EventDisconnect, func(ep *socketio.EventPayload) {
		// Remove the user from the local clients
		socket.RemoveClient(ep.Kws.GetStringAttribute("user_id"))
		fmt.Printf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("user_id"))
	})

	// On close event
	// This event is called when the server disconnects the user actively with .Close() method
	socketio.On(socketio.EventClose, func(ep *socketio.EventPayload) {
		// Remove the user from the local clients
		socket.RemoveClient(ep.Kws.GetStringAttribute("user_id"))
		fmt.Printf("Close event - User: %s", ep.Kws.GetStringAttribute("user_id"))
	})

	// On error event
	socketio.On(socketio.EventError, func(ep *socketio.EventPayload) {
		fmt.Printf("Error event - User: %s", ep.Kws.GetStringAttribute("user_id"))
	})

	app.Get("/ws/:id", socketio.New(func(kws *socketio.Websocket) {
		// Retrieve the user id from endpoint
		userId := kws.Params("id")

		// Add the connection to the list of the connected clients
		// The UUID is generated randomly and is the key that allow
		// socketio to manage Emit/EmitTo/Broadcast
		socket.AddClient(userId, kws.UUID)

		// Every websocket connection has an optional session key => value storage
		kws.SetAttribute("user_id", userId)
		// message := map[string]any{
		// 	"message": "Hello from server",
		// 	"uuid":    kws.UUID,
		// }
		// messageBytes, err := json.Marshal(message)
		// if err != nil {
		// 	fmt.Printf("Failed to marshal message: %+v", err)
		// 	return
		// }
		// kws.EmitTo(kws.UUID, []byte(messageBytes))

		// Broadcast to all the connected users the newcomer
		// kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", userId, kws.UUID)), true)
		// Write welcome message
		kws.Emit([]byte(fmt.Sprintf("Hi!")))
	}))
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
