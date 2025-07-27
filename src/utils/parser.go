package utils

import (
	"github.com/gofiber/fiber/v2"
)

func ParseQueryParams(c *fiber.Ctx, queryParams any) error {
	if err := c.QueryParser(queryParams); err != nil {
		return err
	}
	return nil
}
