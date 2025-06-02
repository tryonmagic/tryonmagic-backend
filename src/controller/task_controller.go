package controller

import (
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/service"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

	"github.com/gofiber/fiber/v2"
)

type TaskController struct {
	TaskService service.TaskService
}

func NewTaskController(taskService service.TaskService) *TaskController {
	return &TaskController{
		TaskService: taskService,
	}
}

// @Summary Create a new task
// @Description Create a new task with the specified type and input data,push it to the queue, and return the task ID.
// @Accept json
// @Produce json
// @Success 200 {object} example.GenericSuccessResponse
// @Failure 500 {object} example.GenericErrorResponse
// @Router /tasks/new [post]
func (t *TaskController) NewTask(c *fiber.Ctx) error {
	req := new(validation.NewTask[any])

	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	task, err := t.TaskService.NewTask(c, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.GenericSuccessResponse{
		Message: "Task created successfully",
		Code:    fiber.StatusCreated,
		Data:    task,
	})
}
