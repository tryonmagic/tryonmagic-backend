package service

import (
	"context"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AIModelService interface {
	CreateAIModel(c *fiber.Ctx, aiModel *model.AIModel) (*model.Task, error)
	BulkCreateAIModels(c *fiber.Ctx, aiModels []model.AIModel) error
	GetAIModels(c *fiber.Ctx, query *validation.BasePaginationQueryParams) (*GetAIModelsResponse, error)
	GetAIModelByID(c *fiber.Ctx, id string) (*model.AIModel, error)
	UpdateAIModel(c *fiber.Ctx, aiModel *model.AIModel) error
	DeleteAIModel(c *fiber.Ctx, id string) error
}

type aIModelService struct {
	Log         *logrus.Logger
	DB          *gorm.DB
	Validate    *validator.Validate
	TaskService TaskService
}

func NewAIModelService(db *gorm.DB, validate *validator.Validate, taskService TaskService) AIModelService {
	return &aIModelService{
		Log:         utils.Log,
		DB:          db,
		Validate:    validate,
		TaskService: taskService,
	}
}

func (s *aIModelService) CreateAIModel(c *fiber.Ctx, aiModel *model.AIModel) (*model.Task, error) {
	userId := c.Locals("user").(*model.User).ID
	aiModel.UserID = &userId
	aiModel.GenAIName = "imagen-3.0-generate-002"
	if userId != uuid.Nil {
		aiModel.AIModelAuthor = model.AIModelAuthorUser
	}

	err := s.DB.WithContext(context.Background()).Create(aiModel).Error
	if err != nil {
		s.Log.Errorf("Failed to create AI model: %+v", err)
		return nil, err
	}
	s.Log.Infof("AI model created successfully: %s", aiModel.ID)
	newTask, err := s.createTask(aiModel, c)
	if err != nil {
		s.Log.Errorf("Failed to create task: %+v", err)
		s.DB.WithContext(context.Background()).Delete(aiModel)
		return nil, err
	}
	s.DB.WithContext(context.Background()).Model(aiModel).Updates(model.AIModel{TaskID: &newTask.ID})
	return newTask, nil
}

func (s *aIModelService) createTask(aiModel *model.AIModel, c *fiber.Ctx) (*model.Task, error) {
	s.Log.Infof("Creating task for AI model: %v", aiModel)
	inputData, err := json.Marshal(aiModel)
	if err != nil {
		s.Log.Errorf("Failed to marshal input data: %+v", err)
		return nil, err
	}
	task := &validation.NewTask[any]{
		Type:      model.TaskTypeGenerateAIModel,
		InputData: datatypes.JSON(inputData),
	}
	s.Log.Infof("task Input data: %v", task)

	newTask, err := s.TaskService.NewTask(c, task)
	if err != nil {
		s.Log.Errorf("Failed to create task: %+v", err)
		return nil, err
	}
	s.Log.Infof("Task created successfully for %v, id: %v", newTask.Type, newTask.ID)
	s.DB.WithContext(context.Background()).Model(aiModel).Updates(model.AIModel{TaskID: &newTask.ID})
	return newTask, nil
}

func (s *aIModelService) BulkCreateAIModels(c *fiber.Ctx, aiModels []model.AIModel) error {
	if len(aiModels) == 0 {
		return nil
	}
	err := s.DB.WithContext(context.Background()).Create(&aiModels).Error
	if err != nil {
		s.Log.Errorf("Failed to bulk create AI models: %+v", err)
		return err
	}
	s.Log.Infof("Bulk AI models created successfully: %d models", len(aiModels))
	for _, aiModel := range aiModels {
		newTask, err := s.createTask(&aiModel, c)
		if err != nil {
			s.Log.Errorf("Failed to create task: %+v", err)
			return err
		}
		s.DB.WithContext(context.Background()).Model(&aiModel).Updates(model.AIModel{TaskID: &newTask.ID})
	}
	return nil
}

func (s *aIModelService) GetAIModels(c *fiber.Ctx, queryParams *validation.BasePaginationQueryParams) (*GetAIModelsResponse, error) {
	var aiModels []model.AIModel
	s.Log.Infof("Query params: %+v", queryParams)

	userId := c.Locals("user").(*model.User).ID
	if userId == uuid.Nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	// Fix: Set the model before counting
	db := s.DB.WithContext(context.Background()).Model(&model.AIModel{}).Where("user_id = ?", userId)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		s.Log.Errorf("Failed to count AI models: %+v", err)
		return nil, err
	}

	offset := GetPaginationOffset(queryParams.Page, queryParams.PageSize)
	if err := db.Offset(offset).Limit(queryParams.PageSize).Order("created_at DESC").Find(&aiModels).Error; err != nil {
		s.Log.Errorf("Failed to get AI models: %+v", err)
		return nil, err
	}
	return &GetAIModelsResponse{
		Data:  aiModels,
		Count: int(count),
	}, nil
}

func (s *aIModelService) GetAIModelByID(c *fiber.Ctx, id string) (*model.AIModel, error) {
	var aiModel model.AIModel
	err := s.DB.WithContext(context.Background()).First(&aiModel, "id = ?", id).Error
	return &aiModel, err
}

func (s *aIModelService) UpdateAIModel(c *fiber.Ctx, aiModel *model.AIModel) error {
	return s.DB.WithContext(context.Background()).Save(aiModel).Error
}

func (s *aIModelService) DeleteAIModel(c *fiber.Ctx, id string) error {
	return s.DB.WithContext(context.Background()).Delete(&model.AIModel{}, "id = ?", id).Error
}
