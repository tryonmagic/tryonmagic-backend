package controller

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mbbipul/prodigyai-backend/src/response"
	"github.com/mbbipul/prodigyai-backend/src/service"

	"github.com/gofiber/fiber/v2"
)

type WebhookController struct {
	WebhookService service.WebhookService
}

func NewWebhookControllerr(webhookService service.WebhookService) *WebhookController {
	return &WebhookController{
		WebhookService: webhookService,
	}
}

func (w *WebhookController) FashnAIVTonWebhook(c *fiber.Ctx) error {

	req := new(service.FashnAIVtonResponseData)

	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	taskId := c.Params("taskID")
	if taskId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "taskId is required")
	}
	// Convert taskId to uuid.UUID
	taskUUID, err := uuid.Parse(taskId)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid taskId format")
	}
	fmt.Printf("FashnAIVTonWebhook taskId: %v", taskUUID)

	err = w.WebhookService.FashnAIVTonWebhook(c, taskUUID, *req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "Webhook processed successfully",
		Code:    fiber.StatusCreated,
		// Data:    task,
	})
}
