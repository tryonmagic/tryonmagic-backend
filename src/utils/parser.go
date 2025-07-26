package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
)

func ParseQueryParams(c *fiber.Ctx, queryParams *validation.BasePaginationQueryParams) error {
	if err := c.QueryParser(queryParams); err != nil {
		return err
	}
	if queryParams.Page == 0 {
		queryParams.Page = 1
	}
	if queryParams.PageSize == 0 {
		queryParams.PageSize = 1
	}
	if queryParams.SortOrder == "" {
		queryParams.SortOrder = validation.SortOrderDesc
	}
	return nil
}
