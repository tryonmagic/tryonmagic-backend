package controller

import (
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/service"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AppInstallationController handles app installation-related API endpoints
type AppInstallationController struct {
	InstallationService service.AppInstallationService
	Validate            *validator.Validate
}

// NewAppInstallationController creates a new app installation controller instance
func NewAppInstallationController(installationService service.AppInstallationService, validate *validator.Validate) *AppInstallationController {
	return &AppInstallationController{
		InstallationService: installationService,
		Validate:            validate,
	}
}

// CheckInstallation checks if a shop has the app installed
// @Summary Check App Installation
// @Description Checks if a shop has the app installed
// @Tags App Installation
// @Produce json
// @Param shopDomain path string true "Shop Domain"
// @Success 200 {object} validation.InstallationCheckResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /installations/{shopDomain} [get]
func (a *AppInstallationController) CheckInstallation(c *fiber.Ctx) error {
	shopDomain := c.Params("shopDomain")
	if shopDomain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Shop domain is required",
			Code:  "MISSING_SHOP_DOMAIN",
		})
	}

	checkResponse, err := a.InstallationService.CheckInstallation(shopDomain)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Installation check completed",
		Code:    fiber.StatusOK,
		Data:    checkResponse,
	})
}

// CreateInstallation creates a new app installation record
// @Summary Create App Installation
// @Description Creates a new app installation record
// @Tags App Installation
// @Accept json
// @Produce json
// @Param request body validation.CreateInstallationRequest true "Installation Data"
// @Success 201 {object} validation.InstallationResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /installations [post]
func (a *AppInstallationController) CreateInstallation(c *fiber.Ctx) error {
	req := new(validation.CreateInstallationRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_BODY",
		})
	}

	if err := a.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "VALIDATION_ERROR",
		})
	}

	createResponse, err := a.InstallationService.CreateInstallation(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "App installation created successfully",
		Code:    fiber.StatusCreated,
		Data:    createResponse,
	})
}

// UpdateInstallation updates an existing app installation
// @Summary Update App Installation
// @Description Updates an existing app installation
// @Tags App Installation
// @Accept json
// @Produce json
// @Param shopDomain path string true "Shop Domain"
// @Param request body validation.UpdateInstallationRequest true "Update Data"
// @Success 200 {object} validation.InstallationResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /installations/{shopDomain} [patch]
func (a *AppInstallationController) UpdateInstallation(c *fiber.Ctx) error {
	shopDomain := c.Params("shopDomain")
	if shopDomain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Shop domain is required",
			Code:  "MISSING_SHOP_DOMAIN",
		})
	}

	req := new(validation.UpdateInstallationRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_BODY",
		})
	}

	if err := a.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "VALIDATION_ERROR",
		})
	}

	updateResponse, err := a.InstallationService.UpdateInstallation(shopDomain, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "App installation updated successfully",
		Code:    fiber.StatusOK,
		Data:    updateResponse,
	})
}

// DeleteInstallation deletes app installation for a shop (called on app/uninstalled webhook)
// @Summary Delete App Installation
// @Description Deletes app installation for a shop (called on app/uninstalled webhook)
// @Tags App Installation
// @Produce json
// @Param shopDomain path string true "Shop Domain"
// @Success 200 {object} validation.InstallationResponse
// @Failure 400 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /installations/{shopDomain} [delete]
func (a *AppInstallationController) DeleteInstallation(c *fiber.Ctx) error {
	shopDomain := c.Params("shopDomain")
	if shopDomain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Shop domain is required",
			Code:  "MISSING_SHOP_DOMAIN",
		})
	}

	deleteResponse, err := a.InstallationService.DeleteInstallation(shopDomain)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "App installation deleted successfully",
		Code:    fiber.StatusOK,
		Data:    deleteResponse,
	})
}

// GetInstallation retrieves an app installation by shop domain
// @Summary Get App Installation
// @Description Retrieves an app installation by shop domain
// @Tags App Installation
// @Produce json
// @Param shopDomain path string true "Shop Domain"
// @Success 200 {object} model.AppInstallation
// @Failure 400 {object} validation.ErrorResponse
// @Failure 404 {object} validation.ErrorResponse
// @Failure 500 {object} validation.ErrorResponse
// @Router /installations/{shopDomain}/details [get]
func (a *AppInstallationController) GetInstallation(c *fiber.Ctx) error {
	shopDomain := c.Params("shopDomain")
	if shopDomain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ErrorResponse{
			Error: "Shop domain is required",
			Code:  "MISSING_SHOP_DOMAIN",
		})
	}

	installation, err := a.InstallationService.GetInstallation(shopDomain)
	if err != nil {
		if err.Error() == "installation not found for shop: "+shopDomain {
			return c.Status(fiber.StatusNotFound).JSON(validation.ErrorResponse{
				Error: "Installation not found",
				Code:  "INSTALLATION_NOT_FOUND",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "App installation retrieved successfully",
		Code:    fiber.StatusOK,
		Data:    installation,
	})
}
