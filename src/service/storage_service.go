package service

import (
	"github.com/mbbipul/prodigyai-backend/src/utils"
	"github.com/mbbipul/prodigyai-backend/src/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StorageService interface {
	SaveFile(c *fiber.Ctx, req *validation.SaveFile) (*model.File, error)
	GetFilesByUserID(c *fiber.Ctx, userID uuid.UUID) ([]*model.File, error)
}

type storageService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewStorageService(db *gorm.DB, validate *validator.Validate) StorageService {
	return &storageService{
		Log:      utils.Log,
		DB:       db,
		Validate: validate,
	}
}

func (s *storageService) SaveFile(c *fiber.Ctx, req *validation.SaveFile) (*model.File, error) {
	s.Log.Infof("Saving file: %+v", req)
	if err := s.Validate.Struct(req); err != nil {
		s.Log.Errorf("Validation error: %+v", err)
		return nil, err
	}
	user := c.Locals("user").(*model.User)
	file := &model.File{
		FullPath: req.FullPath,
		Key:      req.Key,
		URL:      req.URL,
		Name:     req.Name,
		Type:     req.Type,
		Size:     req.Size,
		Width:    req.Width,
		Height:   req.Height,
		Ext:      req.Ext,
		Mime:     req.Mime,
		UserID:   user.ID,
	}
	if err := s.DB.WithContext(c.Context()).Create(file).Error; err != nil {
		s.Log.Errorf("Failed to save file: %+v", err)
		return nil, err
	}
	return file, nil
}

func (s *storageService) GetFilesByUserID(c *fiber.Ctx, userID uuid.UUID) ([]*model.File, error) {
	var files []*model.File
	if err := s.DB.WithContext(c.Context()).Where("user_id = ?", userID).Find(&files).Error; err != nil {
		s.Log.Errorf("Failed to get files: %+v", err)
		return nil, err
	}
	return files, nil
}
