package controller

import (
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/mbbipul/prodigyai-backend/src/response"
	"github.com/mbbipul/prodigyai-backend/src/service"
	"github.com/mbbipul/prodigyai-backend/src/validation"

	"github.com/gofiber/fiber/v2"
)

type StorageController struct {
	StorageService service.StorageService
}

func NewStorageController(storageService service.StorageService) *StorageController {
	return &StorageController{
		StorageService: storageService,
	}
}

// @Summary Save a file
// @Description Save a file with metadata
// @Accept json
// @Produce json
// @Success 200 {object} example.GenericSuccessResponse
// @Failure 500 {object} example.GenericErrorResponse
// @Router /storage/save-file [post]
func (s *StorageController) SaveFile(c *fiber.Ctx) error {
	req := new(validation.SaveFile)

	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	file, err := s.StorageService.SaveFile(c, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "File saved successfully",
		Code:    fiber.StatusCreated,
		Data:    file,
	})
}

// @Summary Get all files
// @Description Get all files for the logged-in user
// @Accept json
// @Produce json
// @Success 200 {object} example.GenericSuccessResponse
// @Failure 500 {object} example.GenericErrorResponse
// @Router /storage/files [get]
func (s *StorageController) GetFiles(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)
	files, err := s.StorageService.GetFilesByUserID(c, user.ID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Files retrieved successfully",
		Code:    fiber.StatusOK,
		Data:    files,
	})
}
