package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/mbbipul/prodigyai-backend/src/utils"
	"github.com/mbbipul/prodigyai-backend/src/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/mbbipul/prodigyai-backend/src/internal/adapter/rabbitmq"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TaskService interface {
	NewTask(c *fiber.Ctx, req any) (*model.Task, error)
}

type taskService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewTaskService(db *gorm.DB, validate *validator.Validate) TaskService {
	return &taskService{
		Log:      utils.Log,
		DB:       db,
		Validate: validate,
	}
}

func (t *taskService) NewTask(c *fiber.Ctx, req any) (*model.Task, error) {
	t.Log.Printf("Creating new task : %v", req)
	newTask, ok := req.(*validation.NewTask[any])
	t.Log.Printf("New task: %v", newTask)
	if !ok {
		t.Log.Errorf("Validation error")
		return nil, errors.New("invalid new task format")
	}

	if err := t.Validate.Struct(newTask); err != nil {
		t.Log.Errorf("Validation error: %+v", err)
		return nil, err
	}
	t.Log.Printf("Creating new task of type %v, input: %v", newTask.Type, newTask.InputData)
	if newTask.Type == model.TaskTypeRemoveBackground {

		// Unmarshal InputData into RemoveBackgroundTaskInput
		var inputData validation.RemoveBackgroundTaskInput
		inputBytes, err := json.Marshal(newTask.InputData)
		if err != nil {
			t.Log.Errorf("Failed to marshal input data: %+v", err)
			return nil, errors.New("invalid input data format")
		}
		if err := json.Unmarshal(inputBytes, &inputData); err != nil {
			t.Log.Errorf("Failed to unmarshal input data into RemoveBackgroundTaskInput: %+v", err)
			return nil, errors.New("invalid input data structure")
		}

		return t.NewRemoveBackgroundTask(c, &validation.NewTask[validation.RemoveBackgroundTaskInput]{
			Type:      newTask.Type,
			InputData: inputData,
		})
	} else if newTask.Type == model.TaskTypeReplaceBackground {
		// Unmarshal InputData into ReplaceBackgroundTaskInput
		var inputData validation.ReplaceBackgroundTaskInput
		inputBytes, err := json.Marshal(newTask.InputData)
		if err != nil {
			t.Log.Errorf("Failed to marshal input data: %+v", err)
			return nil, errors.New("invalid input data format")
		}
		if err := json.Unmarshal(inputBytes, &inputData); err != nil {
			t.Log.Errorf("Failed to unmarshal input data into ReplaceBackgroundTaskInput: %+v", err)
			return nil, errors.New("invalid input data structure")
		}

		return t.NewReplaceBackgroundTask(c, &validation.NewTask[validation.ReplaceBackgroundTaskInput]{
			Type:      newTask.Type,
			InputData: inputData,
		})
	} else if newTask.Type == model.TaskTypeGenerateAIModel {
		var inputData validation.GenerateAIModelTaskInput
		inputBytes, err := json.Marshal(newTask.InputData)
		if err != nil {
			t.Log.Errorf("Failed to marshal input data: %+v", err)
			return nil, errors.New("invalid input data format")
		}
		if err := json.Unmarshal(inputBytes, &inputData); err != nil {
			t.Log.Errorf("Failed to unmarshal input data into GenerateAIModelTaskInput: %+v", err)
			return nil, errors.New("invalid input data structure")
		}

		return t.NewGenerateAIModelTask(c, &validation.NewTask[validation.GenerateAIModelTaskInput]{
			Type:      newTask.Type,
			InputData: inputData,
		})
	}

	return nil, errors.New("unsupported task type")
}

func (t *taskService) NewRemoveBackgroundTask(c *fiber.Ctx, newTask *validation.NewTask[validation.RemoveBackgroundTaskInput]) (*model.Task, error) {

	if err := t.Validate.Struct(newTask); err != nil {
		t.Log.Errorf("Validation error: %+v", err)
		return nil, err
	}

	inputData, err := json.Marshal(newTask.InputData)
	if err != nil {
		t.Log.Errorf("Failed to marshal input data: %+v", err)
		return nil, err
	}
	task := &model.Task{
		Type:      model.TaskTypeRemoveBackground,
		InputData: datatypes.JSON(inputData),
		UserID:    c.Locals("user").(*model.User).ID,
	}

	if err := t.DB.WithContext(c.Context()).Create(task).Error; err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}
	publisher, err := rabbitmq.RemoveBackgroundTaskPublisher()
	if err != nil {
		t.Log.Errorf("Failed to create publisher: %+v", err)
		return nil, err
	}
	err = t.PublishTask(c, publisher, task)
	if err != nil {
		t.Log.Errorf("Failed to publish task: %+v", err)
		return nil, err
	}

	taskID := task.ID

	go func(id uuid.UUID) {
		// Create a new background context
		ctx := context.Background()

		// Update status in a separate transaction
		err := t.DB.WithContext(ctx).
			Model(&model.Task{}).
			Where("id = ?", id).
			Update("status", model.TaskStatusQueued).
			Error

		if err != nil {
			t.Log.Errorf("Failed to update task status: %+v", err)
		}
	}(taskID)

	return task, nil
}

func (t *taskService) NewReplaceBackgroundTask(c *fiber.Ctx, newTask *validation.NewTask[validation.ReplaceBackgroundTaskInput]) (*model.Task, error) {

	t.Log.Printf("Creating NewReplaceBackgroundTask")
	if err := t.Validate.Struct(newTask); err != nil {
		t.Log.Errorf("Validation error: %+v", err)
		return nil, err
	}

	inputData, err := json.Marshal(newTask.InputData)
	if err != nil {
		t.Log.Errorf("Failed to marshal input data: %+v", err)
		return nil, err
	}
	task := &model.Task{
		Type:      newTask.Type,
		InputData: datatypes.JSON(inputData),
		UserID:    c.Locals("user").(*model.User).ID,
	}

	if err := t.DB.WithContext(c.Context()).Create(task).Error; err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}
	publisher, err := rabbitmq.ReplaceBackgroundTaskPublisher()
	if err != nil {
		t.Log.Errorf("Failed to create publisher: %+v", err)
		return nil, err
	}
	err = t.PublishTask(c, publisher, task)
	if err != nil {
		t.Log.Errorf("Failed to publish task: %+v", err)
		return nil, err
	}

	taskID := task.ID

	go func(id uuid.UUID) {
		// Create a new background context
		ctx := context.Background()

		// Update status in a separate transaction
		err := t.DB.WithContext(ctx).
			Model(&model.Task{}).
			Where("id = ?", id).
			Update("status", model.TaskStatusQueued).
			Error

		if err != nil {
			t.Log.Errorf("Failed to update task status: %+v", err)
		}
	}(taskID)

	return task, nil
}

func (t *taskService) NewGenerateAIModelTask(c *fiber.Ctx, newTask *validation.NewTask[validation.GenerateAIModelTaskInput]) (*model.Task, error) {
	t.Log.Printf("Creating NewGenerateAIModelTask")
	if err := t.Validate.Struct(newTask); err != nil {
		t.Log.Errorf("Validation error: %+v", err)
		return nil, err
	}

	inputData, err := json.Marshal(newTask.InputData)
	if err != nil {
		t.Log.Errorf("Failed to marshal input data: %+v", err)
		return nil, err
	}
	task := &model.Task{
		Type:      newTask.Type,
		InputData: datatypes.JSON(inputData),
		UserID:    c.Locals("user").(*model.User).ID,
	}

	if err := t.DB.WithContext(c.Context()).Create(task).Error; err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}
	publisher, err := rabbitmq.GenerateAIModelTaskPublisher()
	if err != nil {
		t.Log.Errorf("Failed to create publisher: %+v", err)
		return nil, err
	}
	err = t.PublishTask(c, publisher, task)
	if err != nil {
		t.Log.Errorf("Failed to publish task: %+v", err)
		return nil, err
	}

	taskID := task.ID

	go func(id uuid.UUID) {
		// Create a new background context
		ctx := context.Background()

		// Update status in a separate transaction
		err := t.DB.WithContext(ctx).
			Model(&model.Task{}).
			Where("id = ?", id).
			Update("status", model.TaskStatusQueued).
			Error

		if err != nil {
			t.Log.Errorf("Failed to update task status: %+v", err)
		}
	}(taskID)

	return task, nil
}

func (t *taskService) PublishTask(c *fiber.Ctx, publisher rabbitmq.BasePublisher, task *model.Task) error {

	headers := rabbitmq.Headers{
		TaskType: string(task.Type),
		JobID:    task.ID.String(),
		Priority: 0,
	}

	newTask := map[string]any{
		"id":        task.ID,
		"inputData": task.InputData,
		"type":      task.Type,
		"createdAt": task.CreatedAt,
		"userId":    task.UserID,
		"shop":      c.Locals("shop").(string),
	}

	t.Log.Infof("Publishing task: %s", newTask)

	err := publisher.Publish(c.Context(), newTask, headers)

	if err != nil {
		t.Log.Errorf("Failed to publish task: %+v", err)
		return err
	}
	t.Log.Infof("Task published successfully: %s", task.ID)
	return nil

}
