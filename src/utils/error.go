package utils

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/prodigyai-backend/src/response"
	"github.com/mbbipul/prodigyai-backend/src/validation"
	"github.com/ztrue/tracerr"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	fmt.Printf("Error: %v", err)
	wrappedErr := tracerr.Wrap(err)
	tracerr.PrintSourceColor(wrappedErr)
	if errorsMap := validation.CustomErrorMessages(err); len(errorsMap) > 0 {
		return response.Error(c, fiber.StatusBadRequest, "Bad Request", errorsMap)
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return response.Error(c, fiberErr.Code, fiberErr.Message, nil)
	}

	return response.Error(c, fiber.StatusInternalServerError, "Internal Server Error", nil)
}

func NotFoundHandler(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusNotFound, "Endpoint Not Found", nil)
}
