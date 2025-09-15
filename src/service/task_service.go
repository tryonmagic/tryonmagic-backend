package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/config"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/pkg/cloudinaryapi"
	"github.com/mbbipul/tryonmagic-backend/src/pkg/downloader"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/mbbipul/tryonmagic-backend/src/internal/adapter/rabbitmq"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TaskService interface {
	NewTask(c *fiber.Ctx, req any) (*model.Task, error)
}

type taskService struct {
	Log                 *logrus.Logger
	DB                  *gorm.DB
	Validate            *validator.Validate
	SubscriptionService SubscriptionService
}

func NewTaskService(db *gorm.DB, subscriptionService SubscriptionService, validate *validator.Validate) TaskService {
	return &taskService{
		Log:                 utils.Log,
		DB:                  db,
		Validate:            validate,
		SubscriptionService: subscriptionService,
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
	switch newTask.Type {
	case model.TaskTypeRemoveBackground:

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
	case model.TaskTypeReplaceBackground:
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
	case model.TaskTypeGenerateAIModel:
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
	case model.TaskTypeVTon:
		var inputData validation.VTONTaskInput
		inputBytes, err := json.Marshal(newTask.InputData)
		if err != nil {
			t.Log.Errorf("Failed to marshal input data: %+v", err)
			return nil, errors.New("invalid input data format")
		}
		if err := json.Unmarshal(inputBytes, &inputData); err != nil {
			t.Log.Errorf("Failed to unmarshal input data into VTONTaskInput: %+v", err)
			return nil, errors.New("invalid input data structure")
		}
		return t.NewVTONTask(c, &validation.NewTask[validation.VTONTaskInput]{
			Type:      newTask.Type,
			InputData: inputData,
		})
	case model.TaskTypeUpscale, model.TaskTypeEnhance:
		var inputData validation.UpscaleTaskInput
		inputBytes, err := json.Marshal(newTask.InputData)
		if err != nil {
			t.Log.Errorf("Failed to marshal input data: %+v", err)
			return nil, errors.New("invalid input data format")
		}
		if err := json.Unmarshal(inputBytes, &inputData); err != nil {
			t.Log.Errorf("Failed to unmarshal input data into UpscaleTaskInput: %+v", err)
			return nil, errors.New("invalid input data structure")
		}
		return t.NewUpscaleTask(c, &validation.NewTask[validation.UpscaleTaskInput]{
			Type:      newTask.Type,
			InputData: inputData,
		})
	case model.TaskTypeExpandImage:
		var inputData validation.ExpandImageTaskInput
		inputBytes, err := json.Marshal(newTask.InputData)
		if err != nil {
			t.Log.Errorf("Failed to marshal input data: %+v", err)
			return nil, errors.New("invalid input data format")
		}
		if err := json.Unmarshal(inputBytes, &inputData); err != nil {
			t.Log.Errorf("Failed to unmarshal input data into ExpandImageTaskInput: %+v", err)
			return nil, errors.New("invalid input data structure")
		}
		return t.NewExpandImageTask(c, &validation.NewTask[validation.ExpandImageTaskInput]{
			Type:      newTask.Type,
			InputData: inputData,
		})
	case model.TaskTypeEditImage:
		var inputData validation.EditImageTaskInput
		inputBytes, err := json.Marshal(newTask.InputData)
		if err != nil {
			t.Log.Errorf("Failed to marshal input data: %+v", err)
			return nil, errors.New("invalid input data format")
		}
		if err := json.Unmarshal(inputBytes, &inputData); err != nil {
			t.Log.Errorf("Failed to unmarshal input data into EditImageTaskInput: %+v", err)
			return nil, errors.New("invalid input data structure")
		}
		return t.NewEditImageTask(c, &validation.NewTask[validation.EditImageTaskInput]{
			Type:      newTask.Type,
			InputData: inputData,
		})
	case model.TaskTypeProductImageEdit:
		var inputData validation.ProductImageEditTaskInput
		inputBytes, err := json.Marshal(newTask.InputData)
		if err != nil {
			t.Log.Errorf("Failed to marshal input data: %+v", err)
			return nil, errors.New("invalid input data format")
		}
		if err := json.Unmarshal(inputBytes, &inputData); err != nil {
			t.Log.Errorf("Failed to unmarshal input data into ProductImageEditTaskInput: %+v", err)
			return nil, errors.New("invalid input data structure")
		}
		return t.NewProductImageEditTask(c, &validation.NewTask[validation.ProductImageEditTaskInput]{
			Type:      newTask.Type,
			InputData: inputData,
		})
	}

	return nil, errors.New("unsupported task type")
}

func (t *taskService) createTask(c *fiber.Ctx, newTask any, taskType model.TaskType, publisherFunc func() (rabbitmq.BasePublisher, error)) (*model.Task, error) {
	t.Log.Printf("Creating task of type %v", taskType)

	if err := t.Validate.Struct(newTask); err != nil {
		t.Log.Errorf("Validation error: %+v", err)
		return nil, err
	}

	inputData, err := json.Marshal(newTask)
	if err != nil {
		t.Log.Errorf("Failed to marshal input data: %+v", err)
		return nil, err
	}
	task := &model.Task{
		Type:      taskType,
		InputData: datatypes.JSON(inputData),
		UserID:    c.Locals("user").(*model.User).ID,
	}

	if err := t.DB.WithContext(context.Background()).Create(task).Error; err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}

	// Deduct points for the task
	if t.SubscriptionService != nil {
		requiredPoints := task.GetPointsForTaskType()
		if err := t.SubscriptionService.DeductPoints(task.UserID, task.ID, requiredPoints, taskType); err != nil {
			t.Log.Errorf("Failed to deduct points: %+v", err)
			// Delete the task if point deduction fails
			if deleteErr := t.DB.WithContext(context.Background()).Delete(task).Error; deleteErr != nil {
				t.Log.Errorf("Failed to delete task after point deduction failure: %+v", deleteErr)
			}
			return nil, err
		}
		t.Log.Infof("Deducted %d points for task %s", requiredPoints, task.ID)
	}

	if publisherFunc != nil {
		publisher, err := publisherFunc()
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
			ctx := context.Background()
			err := t.DB.WithContext(ctx).
				Model(&model.Task{}).
				Where("id = ?", id).
				Update("status", model.TaskStatusQueued).
				Error

			if err != nil {
				t.Log.Errorf("Failed to update task status: %+v", err)
			}
		}(taskID)
	}

	return task, nil
}

func (t *taskService) NewRemoveBackgroundTask(c *fiber.Ctx, newTask *validation.NewTask[validation.RemoveBackgroundTaskInput]) (*model.Task, error) {
	return t.createTask(c, newTask.InputData, model.TaskTypeRemoveBackground, rabbitmq.RemoveBackgroundTaskPublisher)
}

func (t *taskService) NewReplaceBackgroundTask(c *fiber.Ctx, newTask *validation.NewTask[validation.ReplaceBackgroundTaskInput]) (*model.Task, error) {
	return t.createTask(c, newTask.InputData, model.TaskTypeReplaceBackground, rabbitmq.ReplaceBackgroundTaskPublisher)
}

func (t *taskService) NewGenerateAIModelTask(c *fiber.Ctx, newTask *validation.NewTask[validation.GenerateAIModelTaskInput]) (*model.Task, error) {
	return t.createTask(c, newTask.InputData, model.TaskTypeGenerateAIModel, rabbitmq.GenerateAIModelTaskPublisher)
}

func (t *taskService) NewUpscaleTask(c *fiber.Ctx, newTask *validation.NewTask[validation.UpscaleTaskInput]) (*model.Task, error) {
	task, err := t.createTask(c, newTask.InputData, newTask.Type, nil)
	if err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}

	ctx := c.Context()
	sourceImageUrl, err := downloader.GetImageFromSource(ctx, newTask.InputData.SourceImageUrl)
	if err != nil {
		t.Log.Errorf("Failed to get image URL for source image: %v", err)
		return nil, err
	}

	webhookUrl := config.WebhookBaseURL + "/cloudinary/" + config.WebhookAPIKey + "/" + task.ID.String()
	fmt.Println("Webhook URL:", webhookUrl)
	transformation := "e_upscale"
	if newTask.Type == model.TaskTypeEnhance {
		transformation = "e_enhance"
	}

	location, err := cloudinaryapi.UploadImageToCloudinary(sourceImageUrl, task.ID.String(), transformation, webhookUrl, ctx)
	if err != nil {
		t.Log.Errorf("Failed to upload image to Cloudinary: %+v", err)
		return nil, err
	}
	fmt.Println("Cloudinary location:", *location)
	return task, nil
}

func (t *taskService) NewVTONTask(c *fiber.Ctx, newTask *validation.NewTask[validation.VTONTaskInput]) (*model.Task, error) {
	t.Log.Printf("Creating task of type %v", newTask.Type)

	if err := t.Validate.Struct(newTask); err != nil {
		t.Log.Errorf("Validation error: %+v", err)
		return nil, err
	}

	inputData, err := json.Marshal(newTask)
	if err != nil {
		t.Log.Errorf("Failed to marshal input data: %+v", err)
		return nil, err
	}
	task := &model.Task{
		Type:      newTask.Type,
		InputData: datatypes.JSON(inputData),
		UserID:    c.Locals("user").(*model.User).ID,
	}

	if err := t.DB.WithContext(context.Background()).Create(task).Error; err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}

	// Deduct points for the VTON task
	if t.SubscriptionService != nil {
		requiredPoints := task.GetPointsForTaskType()
		if err := t.SubscriptionService.DeductPoints(task.UserID, task.ID, requiredPoints, task.Type); err != nil {
			t.Log.Errorf("Failed to deduct points: %+v", err)
			// Delete the task if point deduction fails
			if deleteErr := t.DB.WithContext(context.Background()).Delete(task).Error; deleteErr != nil {
				t.Log.Errorf("Failed to delete task after point deduction failure: %+v", deleteErr)
			}
			return nil, err
		}
		t.Log.Infof("Deducted %d points for VTON task %s", requiredPoints, task.ID)
	}

	vton, err := GetVtonService(config.VTON_SERVICE)
	if err != nil {
		t.Log.Errorf("Failed to get VTON service: %+v", err)
		return nil, err
	}
	responseData, err := vton.VirtualTryOn(c.Context(), task.ID, newTask.InputData)
	if err != nil {
		t.Log.Errorf("VTON failed: %+v", err)
		if err := t.DB.WithContext(context.Background()).Model(task).Where("id = ?", task.ID).Updates(model.Task{
			Status: model.TaskStatusFailed,
		}).Error; err != nil {
			t.Log.Errorf("Failed to update task status: %+v", err)
		}
		return nil, err
	}
	jobID := responseData.ID

	resultMap := map[string]interface{}{
		"jobId":          jobID,
		"status":         responseData.Status,
		"output":         responseData.Output,
		"outputImageUrl": responseData.Output,
		"error":          responseData.Error,
	}

	resultBytes, err := json.Marshal(resultMap)
	if err != nil {
		if err := t.DB.WithContext(context.Background()).Model(task).Where("id = ?", task.ID).Updates(model.Task{
			Status: model.TaskStatusFailed,
		}).Error; err != nil {
			t.Log.Errorf("Failed to update task status: %+v", err)
		}
		return nil, fmt.Errorf("failed to marshal result data: %w", err)
	}

	resultData := (*datatypes.JSON)(&resultBytes)

	taskObj := new(model.Task)
	taskObj.ID = task.ID
	taskObj.ResultData = resultData
	now := time.Now()
	taskObj.StartedAt = &now

	if err := vton.PostProcessTask(taskObj); err != nil {
		t.Log.Errorf("Failed to post process task: %+v", err)
		return nil, err
	}

	if err := t.DB.WithContext(context.Background()).Model(taskObj).Clauses(clause.Returning{}).Updates(taskObj).Error; err != nil {
		fmt.Printf("Failed to update task: %+v", err)
		return nil, err
	}

	return taskObj, nil
}

func (t *taskService) NewExpandImageTask(c *fiber.Ctx, newTask *validation.NewTask[validation.ExpandImageTaskInput]) (*model.Task, error) {
	return t.createTask(c, newTask.InputData, model.TaskTypeExpandImage, rabbitmq.ExpandImageTaskPublisher)
}

func (t *taskService) NewEditImageTask(c *fiber.Ctx, newTask *validation.NewTask[validation.EditImageTaskInput]) (*model.Task, error) {
	if len(newTask.InputData.EditRegion) == 0 {
		return t.NewEditImageTaskCloudinary(c, newTask)
	}
	return t.createTask(c, newTask.InputData, model.TaskTypeEditImage, rabbitmq.EditImageTaskPublisher)
}

func (t *taskService) NewProductImageEditTask(c *fiber.Ctx, newTask *validation.NewTask[validation.ProductImageEditTaskInput]) (*model.Task, error) {
	if newTask.InputData.ItemsToRecolor == nil || *newTask.InputData.ItemsToRecolor == "" {
		return t.createTask(c, newTask.InputData, model.TaskTypeProductImageEdit, rabbitmq.ProductImageEditTaskPublisher)
	}
	if newTask.InputData.Color == nil || *newTask.InputData.Color == "" {
		return nil, errors.New("color is required")
	}
	return t.NewProductRecolorTask(c, newTask)
}

func (t *taskService) NewProductRecolorTask(c *fiber.Ctx, newTask *validation.NewTask[validation.ProductImageEditTaskInput]) (*model.Task, error) {
	task, err := t.createTask(c, newTask.InputData, newTask.Type, nil)
	if err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}

	ctx := c.Context()
	sourceImageUrl, err := downloader.GetImageFromSource(ctx, newTask.InputData.SourceImageUrl)
	if err != nil {
		t.Log.Errorf("Failed to get image URL for source image: %v", err)
		return nil, err
	}

	webhookUrl := config.WebhookBaseURL + "/cloudinary/" + config.WebhookAPIKey + "/" + task.ID.String()
	fmt.Println("Webhook URL:", webhookUrl)
	itemsToRecolor := strings.TrimSpace(*newTask.InputData.ItemsToRecolor)
	itemsToRecolor = strings.Join(strings.Split(itemsToRecolor, ","), ";")

	transformation := fmt.Sprintf("e_gen_recolor:prompt_%s;to-color_%s;multiple_true", itemsToRecolor, *newTask.InputData.Color)
	location, err := cloudinaryapi.UploadImageToCloudinary(sourceImageUrl, task.ID.String(), transformation, webhookUrl, ctx)
	if err != nil {
		t.Log.Errorf("Failed to upload image to Cloudinary: %+v", err)
		return nil, err
	}
	fmt.Println("Cloudinary location:", *location)
	return task, nil
}

func (t *taskService) NewEditImageTaskCloudinary(c *fiber.Ctx, newTask *validation.NewTask[validation.EditImageTaskInput]) (*model.Task, error) {
	task, err := t.createTask(c, newTask.InputData, newTask.Type, nil)
	if err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}
	ctx := c.Context()
	sourceImageUrl, err := downloader.GetImageFromSource(ctx, newTask.InputData.SourceImageUrl)
	if err != nil {
		t.Log.Errorf("Failed to get image URL for source image: %v", err)
		return nil, err
	}

	webhookUrl := config.WebhookBaseURL + "/cloudinary/" + config.WebhookAPIKey + "/" + task.ID.String() + "?taskType=" + string(newTask.Type)
	fmt.Println("Webhook URL:", webhookUrl)
	var transformation string
	switch newTask.InputData.EditType {
	case validation.EditTypeInpaintRemove:
		transformation = fmt.Sprintf("e_gen_remove:prompt_%s;multiple_true;remove-shadow_true", *newTask.InputData.EditPrompt)
	case validation.EditTypeInpaintReplace:
		from := strings.TrimSpace(*newTask.InputData.ReplacePrompt.ReplaceFrom)
		to := strings.TrimSpace(newTask.InputData.ReplacePrompt.ReplaceTo)
		if from == "" {
			return nil, errors.New("replaceFrom is required")
		} else if to == "" {
			return nil, errors.New("replaceTo is required")
		}
		transformation = fmt.Sprintf("e_gen_replace:from_%s;to_%s;multiple_true", from, to)
	}
	if transformation == "" {
		return nil, errors.New("invalid edit type")
	}

	location, err := cloudinaryapi.UploadImageToCloudinary(sourceImageUrl, task.ID.String(), transformation, webhookUrl, ctx)
	if err != nil {
		t.Log.Errorf("Failed to upload image to Cloudinary: %+v", err)
		return nil, err
	}
	fmt.Println("Cloudinary location:", *location)
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
