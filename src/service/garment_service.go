package service

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/mbbipul/prodigyai-backend/src/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type GarmentService interface {
	GetGarments(c *fiber.Ctx) ([]model.Garment, error)
	CreateGarment(c *fiber.Ctx, garment *model.Garment) error
	GetGarmentByID(c *fiber.Ctx, id string) (*model.Garment, error)
	UpdateGarment(c *fiber.Ctx, garment *model.Garment) error
	DeleteGarment(c *fiber.Ctx, id string) error
}

type garmentService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewGarmentService(db *gorm.DB, validate *validator.Validate) GarmentService {
	return &garmentService{
		Log:      utils.Log,
		DB:       db,
		Validate: validate,
	}
}

func (s *garmentService) GetGarments(c *fiber.Ctx) ([]model.Garment, error) {
	var garments []model.Garment

	userId := c.Locals("user").(*model.User).ID
	if userId == uuid.Nil {
		return garments, nil
	}
	err := s.DB.WithContext(c.Context()).Find(&garments).Where(&model.Garment{UserID: &userId}).Order("created_at DESC").Error
	return garments, err
}

func (s *garmentService) CreateGarment(c *fiber.Ctx, garment *model.Garment) error {
	userId := c.Locals("user").(*model.User).ID
	garment.UserID = &userId
	return s.DB.Create(garment).Error
}

func (s *garmentService) GetGarmentByID(c *fiber.Ctx, id string) (*model.Garment, error) {
	var garment model.Garment
	if err := s.DB.First(&garment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &garment, nil
}

func (s *garmentService) UpdateGarment(c *fiber.Ctx, garment *model.Garment) error {
	return s.DB.WithContext(c.Context()).Save(garment).Error
}

func (s *garmentService) DeleteGarment(c *fiber.Ctx, id string) error {
	return s.DB.Delete(&model.Garment{}, "id = ?", id).Error
}
