package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/mbbipul/prodigyai-backend/src/response"
	"github.com/mbbipul/prodigyai-backend/src/service"
)

type AIModelController struct {
	AIModelService service.AIModelService
}

func NewAIModelController(aiModelService service.AIModelService) *AIModelController {
	return &AIModelController{AIModelService: aiModelService}
}

// @Tags         AI Models
// @Summary      Create an AI model
// @Description  Create a new AI model.
// @Security     BearerAuth
// @Produce      json
// @Param        request  body  model.AIModel  true  "AI Model data"
// @Router       /ai-models [post]
// @Success      201  {object}  response.GenericSuccessResponse
// @Failure      400  {object}  response.GenericErrorResponse  "Invalid request body"
// @Failure      500  {object}  response.GenericErrorResponse  "Internal server error"
func (c *AIModelController) CreateAIModel(ctx *fiber.Ctx) error {
	var aiModel model.AIModel
	if err := ctx.BodyParser(&aiModel); err != nil {
		return err
	}
	if err := c.AIModelService.CreateAIModel(ctx, &aiModel); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to create AI model",
			Errors:  nil,
		})
	}
	return ctx.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "AI model created successfully",
		Code:    fiber.StatusCreated,
		Data:    aiModel,
	})
}

// @Tags         AI Models
// @Summary      Get all AI models
// @Description  Retrieve a list of all AI models.
// @Security     BearerAuth
// @Produce      json
// @Router       /ai-models [get]
// @Success      200  {object}  response.GenericSuccessResponse
// @Failure      500  {object}  response.GenericErrorResponse  "Internal server error"
func (c *AIModelController) GetAIModels(ctx *fiber.Ctx) error {
	aiModels, err := c.AIModelService.GetAIModels(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to fetch AI models",
			Errors:  nil,
		})
	}
	return ctx.JSON(response.GenericSuccessResponse{
		Message: "AI models fetched successfully",
		Code:    fiber.StatusOK,
		Data:    aiModels,
	})
}

// @Tags         AI Models
// @Summary      Get an AI model by ID
// @Description  Retrieve a specific AI model by its ID.
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "AI Model ID"
// @Router       /ai-models/{id} [get]
// @Success      200  {object}  response.GenericSuccessResponse
// @Failure      400  {object}  response.GenericErrorResponse  "Invalid ID"
// @Failure      404  {object}  response.GenericErrorResponse  "AI model not found"
func (c *AIModelController) GetAIModelByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	aiModel, err := c.AIModelService.GetAIModelByID(ctx, id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusNotFound,
			Message: "AI model not found",
			Errors:  nil,
		})
	}
	return ctx.JSON(response.GenericSuccessResponse{
		Message: "AI model fetched successfully",
		Code:    fiber.StatusOK,
		Data:    aiModel,
	})
}

// @Tags         AI Models
// @Summary      Update an AI model
// @Description  Update an existing AI model by its ID.
// @Security     BearerAuth
// @Produce      json
// @Param        id       path  string       true  "AI Model ID"
// @Param        request  body  model.AIModel  true  "Updated AI Model data"
// @Router       /ai-models/{id} [patch]
// @Success      200  {object}  response.GenericSuccessResponse
// @Failure      400  {object}  response.GenericErrorResponse  "Invalid request body or ID"
// @Failure      500  {object}  response.GenericErrorResponse  "Internal server error"
func (c *AIModelController) UpdateAIModel(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var aiModel model.AIModel
	if err := ctx.BodyParser(&aiModel); err != nil {
		return err
	}
	modelId, err := uuid.Parse(id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid user ID",
			Errors:  nil,
		})
	}
	aiModel.ID = modelId
	if err := c.AIModelService.UpdateAIModel(ctx, &aiModel); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to update AI model",
			Errors:  nil,
		})
	}
	return ctx.JSON(response.GenericSuccessResponse{
		Message: "AI model updated successfully",
		Code:    fiber.StatusOK,
		Data:    aiModel,
	})
}

// @Tags         AI Models
// @Summary      Delete an AI model
// @Description  Delete an AI model by its ID.
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "AI Model ID"
// @Router       /ai-models/{id} [delete]
// @Success      204  {object}  response.GenericSuccessResponse
// @Failure      500  {object}  response.GenericErrorResponse  "Internal server error"
func (c *AIModelController) DeleteAIModel(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.AIModelService.DeleteAIModel(ctx, id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to delete AI model",
			Errors:  nil,
		})
	}
	return ctx.Status(fiber.StatusNoContent).JSON(response.GenericSuccessResponse{
		Message: "AI model deleted successfully",
		Code:    fiber.StatusNoContent,
	})
}

// @Tags         AI Models
// @Summary      Bulk create AI models
// @Description  Create multiple AI models at once.
// @Security     BearerAuth
// @Produce      json
// @Param        request  body  []model.AIModel  true  "List of AI Models"
// @Router       /ai-models/bulk [post]
// @Success      201  {object}  response.GenericSuccessResponse
// @Failure      400  {object}  response.GenericErrorResponse  "Invalid request body"
// @Failure      500  {object}  response.GenericErrorResponse  "Internal server error"
func (c *AIModelController) BulkCreateAIModels(ctx *fiber.Ctx) error {
	var aiModels []model.AIModel
	if err := ctx.BodyParser(&aiModels); err != nil {
		return err
	}
	if err := c.AIModelService.BulkCreateAIModels(ctx, aiModels); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to create AI models",
			Errors:  nil,
		})
	}
	return ctx.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "AI models created successfully",
		Code:    fiber.StatusCreated,
		Data:    aiModels,
	})
}
