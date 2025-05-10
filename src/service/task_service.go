package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mbbipul/prodigyai-backend/src/config"
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/mbbipul/prodigyai-backend/src/pkg/cloudinaryapi"
	"github.com/mbbipul/prodigyai-backend/src/pkg/downloader"
	"github.com/mbbipul/prodigyai-backend/src/utils"
	"github.com/mbbipul/prodigyai-backend/src/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/mbbipul/prodigyai-backend/src/internal/adapter/rabbitmq"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	} else if newTask.Type == model.TaskTypeVTon {
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
	} else if newTask.Type == model.TaskTypeUpscale {
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

	if err := t.DB.WithContext(c.Context()).Create(task).Error; err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
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
	task, err := t.createTask(c, newTask.InputData, model.TaskTypeUpscale, nil)
	if err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}
	func(taskId uuid.UUID, sourceImage string) {
		ctx := c.Context()
		sourceImageUrl, err := downloader.GetImageFromSource(ctx, sourceImage)
		if err != nil {
			t.Log.Errorf("Failed to get image URL for source image: %v", err)
			return
		}

		webhookUrl := config.WebhookBaseURL + "/cloudinary/" + config.WebhookAPIKey + "/" + taskId.String()
		fmt.Println("Webhook URL:", webhookUrl)

		location, err := cloudinaryapi.UploadImageToCloudinary(sourceImageUrl, taskId.String(), "e_upscale", webhookUrl, ctx)
		if err != nil {
			t.Log.Errorf("Failed to upload image to Cloudinary: %+v", err)
			return
		}
		fmt.Println("Cloudinary location:", *location)

		// downloader := downloader.BuildDownloader(transformedUrl)
		// if downloader == nil {
		// 	return
		// }

		// key := "upscale/" + task.ID.String() + ".png"
		// err = cloudinaryapi.SaveImageToS3(ctime, transformedUrl, config.S3BucketName, key)
		// if err != nil {
		// 	t.Log.Errorf("Failed to upload image to S3: %+v", err)
		// 	return
		// }
		// resultMap := map[string]interface{}{
		// 	"outputImageUrl": "s3://" + config.S3BucketName + "/" + key,
		// }
		// resultBytes, err := json.Marshal(resultMap)
		// if err != nil {
		// 	if err := t.DB.WithContext(ctime).Model(task).Where("id = ?", task.ID).Updates(model.Task{
		// 		Status: model.TaskStatusFailed,
		// 	}).Error; err != nil {
		// 		t.Log.Errorf("Failed to update task status: %+v", err)
		// 	}
		// 	return
		// }

		// resultData := (*datatypes.JSON)(&resultBytes)

		// taskObj := new(model.Task)
		// taskObj.ID = taskId
		// taskObj.ResultData = resultData
		// taskObj.Status = model.TaskStatusSucceeded
		// now := time.Now()
		// taskObj.CompletedAt = &now

		// if err := t.DB.WithContext(context.Background()).Model(taskObj).Clauses(clause.Returning{}).Updates(taskObj).Error; err != nil {
		// 	fmt.Printf("Failed to update task: %+v", err)
		// 	return
		// }
		// err = socket.EmitEventToUser(taskObj.UserID.String(), "task-progress", map[string]any{
		// 	"id":         taskObj.ID,
		// 	"type":       taskObj.Type,
		// 	"status":     taskObj.Status,
		// 	"resultData": taskObj.ResultData,
		// })
		// if err != nil {
		// 	fmt.Printf("Failed to emit event to user: %+v", err)
		// }

	}(task.ID, newTask.InputData.SourceImageUrl)
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

	if err := t.DB.WithContext(c.Context()).Create(task).Error; err != nil {
		t.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}
	fashn := NewFashnAIVirtualTryOn(config.FASHN_AI_API_KEY, t.Log)
	responseData, err := fashn.VirtualTryOn(c.Context(), task.ID, newTask.InputData)
	if err != nil {
		t.Log.Errorf("VTON failed: %+v", err)
		if err := t.DB.WithContext(c.Context()).Model(task).Where("id = ?", task.ID).Updates(model.Task{
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
		if err := t.DB.WithContext(c.Context()).Model(task).Where("id = ?", task.ID).Updates(model.Task{
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

	if err := t.DB.WithContext(context.Background()).Model(taskObj).Clauses(clause.Returning{}).Updates(taskObj).Error; err != nil {
		fmt.Printf("Failed to update task: %+v", err)
		return nil, err
	}

	return taskObj, nil
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
