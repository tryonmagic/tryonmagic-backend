package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/service"
)

type GarmentController struct {
	GarmentService service.GarmentService
}

func NewGarmentController(garmentService service.GarmentService) *GarmentController {
	return &GarmentController{GarmentService: garmentService}
}

func (g *GarmentController) GetGarments(ctx *fiber.Ctx) error {
	garments, err := g.GarmentService.GetGarments(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to fetch Garments",
			Errors:  nil,
		})
	}
	return ctx.JSON(response.GenericSuccessResponse{
		Message: "Garments fetched successfully",
		Code:    fiber.StatusOK,
		Data:    garments,
	})
}

func (g *GarmentController) CreateGarment(c *fiber.Ctx) error {
	var garment model.Garment
	if err := c.BodyParser(&garment); err != nil {
		return err
	}
	if err := g.GarmentService.CreateGarment(c, &garment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to create Garment",
			Errors:  nil,
		})
	}
	return c.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "Garment created successfully",
		Code:    fiber.StatusCreated,
		Data:    garment,
	})
}

func (g *GarmentController) GetGarmentByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	aiModel, err := g.GarmentService.GetGarmentByID(ctx, id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusNotFound,
			Message: "Garment not found",
			Errors:  nil,
		})
	}
	return ctx.JSON(response.GenericSuccessResponse{
		Message: "Garment fetched successfully",
		Code:    fiber.StatusOK,
		Data:    aiModel,
	})
}

func (g *GarmentController) UpdateGarment(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var garment model.Garment
	if err := ctx.BodyParser(&garment); err != nil {
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
	garment.ID = modelId
	if err := g.GarmentService.UpdateGarment(ctx, &garment); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to update Garment",
			Errors:  nil,
		})
	}
	return ctx.JSON(response.GenericSuccessResponse{
		Message: "Garment updated successfully",
		Code:    fiber.StatusOK,
		Data:    garment,
	})
}

func (g *GarmentController) DeleteGarment(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := g.GarmentService.DeleteGarment(ctx, id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.GenericErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to delete Garment",
			Errors:  nil,
		})
	}
	return ctx.Status(fiber.StatusNoContent).JSON(response.GenericSuccessResponse{
		Message: "Garment deleted successfully",
		Code:    fiber.StatusNoContent,
	})
}
