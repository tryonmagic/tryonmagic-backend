package controller

import (
	"fmt"
	"math"

	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/service"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserController struct {
	UserService  service.UserService
	TokenService service.TokenService
}

func NewUserController(userService service.UserService, tokenService service.TokenService) *UserController {
	return &UserController{
		UserService:  userService,
		TokenService: tokenService,
	}
}

// @Tags         Users
// @Summary      Get all users
// @Description  Only admins can retrieve all users.
// @Security BearerAuth
// @Produce      json
// @Param        page     query     int     false   "Page number"  default(1)
// @Param        limit    query     int     false   "Maximum number of users"    default(10)
// @Param        search   query     string  false  "Search by name or email or role"
// @Router       /users [get]
// @Success      200  {object}  example.GetAllUserResponse
// @Failure      401  {object}  example.Unauthorized  "Unauthorized"
// @Failure      403  {object}  example.Forbidden  "Forbidden"
func (u *UserController) GetUsers(c *fiber.Ctx) error {
	query := &validation.QueryUser{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: c.Query("search", ""),
	}

	users, totalResults, err := u.UserService.GetUsers(c, query)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(response.SuccessWithPaginate[model.User]{
			Code:    fiber.StatusOK,
			Status:  "success",
			Message: "Get all users successfully",
			Data: response.PaginateData[model.User]{
				Results:      users,
				Page:         query.Page,
				Limit:        query.Limit,
				TotalPages:   int64(math.Ceil(float64(totalResults) / float64(query.Limit))),
				TotalResults: totalResults,
			},
		})
}

// @Tags         Users
// @Summary      Get a user
// @Description  Logged in users can fetch only their own user information. Only admins can fetch other users.
// @Security BearerAuth
// @Produce      json
// @Param        id  path  string  true  "User id"
// @Router       /users/{id} [get]
// @Success      200  {object}  example.GetUserResponse
// @Failure      401  {object}  example.Unauthorized  "Unauthorized"
// @Failure      403  {object}  example.Forbidden  "Forbidden"
// @Failure      404  {object}  example.NotFound  "Not found"
func (u *UserController) GetUserByID(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		userID = c.Locals("user").(*model.User).ID.String()
	}

	if _, err := uuid.Parse(userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	user, err := u.UserService.GetUserByID(c, userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(response.GenericSuccessResponse{
			Code:    fiber.StatusOK,
			Message: "Get user successfully",
			Data:    user,
		})
}

// @Tags         Users
// @Summary      Create a user
// @Description  Only admins can create other users.
// @Security BearerAuth
// @Produce      json
// @Param        request  body  validation.CreateUser  true  "Request body"
// @Router       /users [post]
// @Success      201  {object}  example.CreateUserResponse
// @Failure      401  {object}  example.Unauthorized  "Unauthorized"
// @Failure      403  {object}  example.Forbidden  "Forbidden"
// @Failure      409  {object}  example.DuplicateEmail  "Email already taken"
func (u *UserController) CreateUser(c *fiber.Ctx) error {
	req := new(validation.CreateUser)

	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := u.UserService.CreateUser(c, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).
		JSON(response.SuccessWithUser{
			Code:    fiber.StatusCreated,
			Status:  "success",
			Message: "Create user successfully",
			User:    *user,
		})
}

// @Tags         Users
// @Summary      Update a user
// @Description  Logged in users can only update their own information. Only admins can update other users.
// @Security BearerAuth
// @Produce      json
// @Param        id  path  string  true  "User id"
// @Param        request  body  validation.UpdateUser  true  "Request body"
// @Router       /users/{id} [patch]
// @Success      200  {object}  example.UpdateUserResponse
// @Failure      401  {object}  example.Unauthorized  "Unauthorized"
// @Failure      403  {object}  example.Forbidden  "Forbidden"
// @Failure      404  {object}  example.NotFound  "Not found"
// @Failure      409  {object}  example.DuplicateEmail  "Email already taken"
func (u *UserController) UpdateUser(c *fiber.Ctx) error {
	req := new(validation.UpdateUser)
	userID := c.Params("userId")

	if _, err := uuid.Parse(userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := u.UserService.UpdateUser(c, req, userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(response.SuccessWithUser{
			Code:    fiber.StatusOK,
			Status:  "success",
			Message: "Update user successfully",
			User:    *user,
		})
}

// @Tags         Users
// @Summary      Delete a user
// @Description  Logged in users can delete only themselves. Only admins can delete other users.
// @Security BearerAuth
// @Produce      json
// @Param        id  path  string  true  "User id"
// @Router       /users/{id} [delete]
// @Success      200  {object}  example.DeleteUserResponse
// @Failure      401  {object}  example.Unauthorized  "Unauthorized"
// @Failure      403  {object}  example.Forbidden  "Forbidden"
// @Failure      404  {object}  example.NotFound  "Not found"
func (u *UserController) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("userId")

	if _, err := uuid.Parse(userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	if err := u.TokenService.DeleteAllToken(c, userID); err != nil {
		return err
	}

	if err := u.UserService.DeleteUser(c, userID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(response.Common{
			Code:    fiber.StatusOK,
			Status:  "success",
			Message: "Delete user successfully",
		})
}

// @Summary Get user usage history
// @Description Get comprehensive usage statistics for the authenticated user
// @Accept json
// @Produce json
// @Success 200 {object} response.GenericSuccessResponse
// @Failure 401 {object} response.GenericErrorResponse
// @Failure 500 {object} response.GenericErrorResponse
// @Router /usage/history [get]
func (u *UserController) GetUsageHistory(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	usageSummary, err := u.UserService.GetUsageHistory(c, user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve usage history")
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Message: "Usage history retrieved successfully",
		Code:    fiber.StatusOK,
		Data:    usageSummary,
	})
}

func (u *UserController) GetTasks(c *fiber.Ctx) error {
	fmt.Println("GetTasks")
	user := c.Locals("user").(*model.User)
	queryParams := &validation.GetTasksQueryParams{}
	if err := utils.ParseQueryParams(c, queryParams); err != nil {
		return err
	}
	if queryParams.Page == 0 {
		queryParams.Page = 1
	}
	if queryParams.PageSize == 0 {
		queryParams.PageSize = 10
	}
	tasks, err := u.UserService.GetTasks(c, user.ID, *queryParams)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessWithPaginate[model.Task]{
		Message: "Tasks retrieved successfully",
		Code:    fiber.StatusOK,
		Data: response.PaginateData[model.Task]{
			Results:      tasks.Results,
			Page:         queryParams.Page,
			Limit:        queryParams.PageSize,
			TotalPages:   service.CalculateTotalPages(int64(tasks.Count), queryParams.PageSize),
			TotalResults: tasks.Count,
		},
	})
}

// @Summary Audit a task
// @Description Report an issue with a task and optionally request a refund
// @Tags Users
// @Accept json
// @Produce json
// @Param taskId path string true "Task ID"
// @Param request body validation.AuditTaskRequest true "Audit request body"
// @Success 200 {object} response.GenericSuccessResponse
// @Security BearerAuth
// @Router /users/tasks/{taskId}/audit [post]
func (u *UserController) AuditTask(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)
	taskIDStr := c.Params("taskId")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	req := new(validation.AuditTaskRequest)
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := u.UserService.AuditTask(c, user.ID, taskID, req); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse{
		Code:    fiber.StatusOK,
		Message: "Task audit reported successfully",
	})
}
