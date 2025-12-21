package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	m "github.com/mbbipul/tryonmagic-backend/src/middleware"
	"github.com/mbbipul/tryonmagic-backend/src/service"
)

func SubscriptionRoutes(
	v1 fiber.Router, a service.AuthService, u service.UserService, s service.SubscriptionService,
) {
	subscriptionController := controller.NewSubscriptionController(s)

	subscription := v1.Group("/subscription")
	subscription.Use(m.Auth(a, u))

	subscription.Get("/status", subscriptionController.GetSubscriptionStatus)

	subscription.Put("/plan", subscriptionController.UpdateSubscriptionPlan)
	subscription.Post("/shopify/sync", subscriptionController.SyncShopifySubscription)

	// Transaction history endpoints
	subscription.Get("/transactions", subscriptionController.GetTransactionHistory)
	subscription.Get("/transactions/type", subscriptionController.GetTransactionHistoryByType)
	subscription.Get("/transactions/summary", subscriptionController.GetTransactionSummary)

	// Purchase endpoints
	subscription.Post("/purchase-points", subscriptionController.PurchasePoints)
	subscription.Post("/purchase", subscriptionController.PurchaseSubscription)

	// Refund points (for failed tasks)
	subscription.Post("/refund", subscriptionController.RefundPoints)

	// Admin routes
	admin := subscription.Group("/admin/subscription")
	admin.Post("/add-points", subscriptionController.AddPoints)
}
