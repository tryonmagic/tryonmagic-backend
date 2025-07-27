package controller

import (
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/service"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

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
	queryParams := &validation.GetFilesQueryParams{}
	if err := utils.ParseQueryParams(c, queryParams); err != nil {
		return err
	}
	if queryParams.Page == 0 {
		queryParams.Page = 1
	}
	if queryParams.PageSize == 0 {
		queryParams.PageSize = 10
	}
	files, err := s.StorageService.GetFilesByUserID(c.Context(), user.ID, queryParams)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessWithPaginate[model.File]{
		Message: "Files retrieved successfully",
		Code:    fiber.StatusOK,
		Data: response.PaginateData[model.File]{
			Results:      files.Results,
			Page:         queryParams.Page,
			Limit:        queryParams.PageSize,
			TotalPages:   service.CalculateTotalPages(int64(files.Count), queryParams.PageSize),
			TotalResults: files.Count,
		},
	})
}
