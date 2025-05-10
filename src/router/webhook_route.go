package router

import (
	"github.com/mbbipul/prodigyai-backend/src/controller"
	m "github.com/mbbipul/prodigyai-backend/src/middleware"
	"github.com/mbbipul/prodigyai-backend/src/service"

	"github.com/gofiber/fiber/v2"
)

func WebhookRoutes(v1 fiber.Router, w service.WebhookService) {
	webhookController := controller.NewWebhookControllerr(w)

	webhooks := v1.Group("/webhooks")

	webhooks.Post("/fashnai-vton/:apiKey/:taskID", m.WebhookAuth(), webhookController.FashnAIVTonWebhook)
	webhooks.Post("/cloudinary/:apiKey/:taskID", m.WebhookAuth(), webhookController.CloudinaryWebhook)

}
