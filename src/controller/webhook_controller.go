package controller

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/pkg/cloudinaryapi"
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/service"

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
	body := c.Body()
	fmt.Printf("FashnAIVTonWebhook body: %v", string(body))

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

func (w *WebhookController) CloudinaryWebhook(c *fiber.Ctx) error {

	req := new(cloudinaryapi.CloudinaryResponse)
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	json, err := json.Marshal(req)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}

	fmt.Printf("CloudinaryWebhook req: %v", string(json))

	taskId := c.Params("taskID")
	if taskId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "taskId is required")
	}
	// Convert taskId to uuid.UUID
	taskUUID, err := uuid.Parse(taskId)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid taskId format")
	}
	fmt.Printf("CloudinaryWebhook taskId: %v", taskUUID)
	taskType := c.Query("taskType")
	err = w.WebhookService.CloudinaryWebhook(c.Context(), taskUUID, *req, taskType)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "Webhook processed successfully",
		Code:    fiber.StatusCreated,
		// Data:    task,
	})
}
