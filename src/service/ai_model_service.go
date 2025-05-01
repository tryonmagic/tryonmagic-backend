package service

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/mbbipul/prodigyai-backend/src/utils"
	"github.com/mbbipul/prodigyai-backend/src/validation"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AIModelService interface {
	CreateAIModel(c *fiber.Ctx, aiModel *model.AIModel) error
	BulkCreateAIModels(c *fiber.Ctx, aiModels []model.AIModel) error
	GetAIModels(c *fiber.Ctx) ([]model.AIModel, error)
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

func (s *aIModelService) CreateAIModel(c *fiber.Ctx, aiModel *model.AIModel) error {
	userId := c.Locals("user").(*model.User).ID
	aiModel.UserID = &userId
	aiModel.GenAIName = "imagen-3.0-generate-002"
	if userId != uuid.Nil {
		aiModel.AIModelAuthor = model.AIModelAuthorUser
	}

	err := s.DB.WithContext(c.Context()).Create(aiModel).Error
	if err != nil {
		s.Log.Errorf("Failed to create AI model: %+v", err)
		return err
	}
	s.Log.Infof("AI model created successfully: %s", aiModel.ID)
	err = s.createTask(aiModel, c)
	if err != nil {
		s.Log.Errorf("Failed to create task: %+v", err)
		return nil
	}
	return nil
}

func (s *aIModelService) createTask(aiModel *model.AIModel, c *fiber.Ctx) error {
	s.Log.Infof("Creating task for AI model: %v", aiModel)
	inputData, err := json.Marshal(aiModel)
	if err != nil {
		s.Log.Errorf("Failed to marshal input data: %+v", err)
		return err
	}
	task := &validation.NewTask[any]{
		Type:      model.TaskTypeGenerateAIModel,
		InputData: datatypes.JSON(inputData),
	}
	s.Log.Infof("task Input data: %v", task)

	newTask, err := s.TaskService.NewTask(c, task)
	if err != nil {
		s.Log.Errorf("Failed to create task: %+v", err)
		return nil
	}
	s.Log.Infof("Task created successfully for %v, id: %v", newTask.Type, newTask.ID)
	s.DB.WithContext(c.Context()).Model(aiModel).Updates(model.AIModel{TaskID: &newTask.ID})
	return nil
}

func (s *aIModelService) BulkCreateAIModels(c *fiber.Ctx, aiModels []model.AIModel) error {
	if len(aiModels) == 0 {
		return nil
	}
	err := s.DB.WithContext(c.Context()).Create(&aiModels).Error
	if err != nil {
		s.Log.Errorf("Failed to bulk create AI models: %+v", err)
		return err
	}
	s.Log.Infof("Bulk AI models created successfully: %d models", len(aiModels))
	for _, aiModel := range aiModels {
		err = s.createTask(&aiModel, c)
		if err != nil {
			s.Log.Errorf("Failed to create task: %+v", err)
			return nil
		}
	}
	return nil
}

func (s *aIModelService) GetAIModels(c *fiber.Ctx) ([]model.AIModel, error) {
	var aiModels []model.AIModel

	userId := c.Locals("user").(*model.User).ID
	if userId == uuid.Nil {
		return aiModels, nil
	}
	err := s.DB.WithContext(c.Context()).Find(&aiModels).Where(&model.AIModel{UserID: &userId}).Order("created_at DESC").Error
	return aiModels, err
}

func (s *aIModelService) GetAIModelByID(c *fiber.Ctx, id string) (*model.AIModel, error) {
	var aiModel model.AIModel
	err := s.DB.WithContext(c.Context()).First(&aiModel, "id = ?", id).Error
	return &aiModel, err
}

func (s *aIModelService) UpdateAIModel(c *fiber.Ctx, aiModel *model.AIModel) error {
	return s.DB.WithContext(c.Context()).Save(aiModel).Error
}

func (s *aIModelService) DeleteAIModel(c *fiber.Ctx, id string) error {
	return s.DB.WithContext(c.Context()).Delete(&model.AIModel{}, "id = ?", id).Error
}
