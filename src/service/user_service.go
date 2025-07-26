package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserService interface {
	GetUsers(c *fiber.Ctx, params *validation.QueryUser) ([]model.User, int64, error)
	GetUserByID(c *fiber.Ctx, id string) (*model.User, error)
	GetUserByEmail(c *fiber.Ctx, email string) (*model.User, error)
	CreateUser(c *fiber.Ctx, req *validation.CreateUser) (*model.User, error)
	UpdatePassOrVerify(c *fiber.Ctx, req *validation.UpdatePassOrVerify, id string) error
	UpdateUser(c *fiber.Ctx, req *validation.UpdateUser, id string) (*model.User, error)
	DeleteUser(c *fiber.Ctx, id string) error
	CreateGoogleUser(c *fiber.Ctx, req *validation.GoogleLogin) (*model.User, error)
	GetUsageHistory(c *fiber.Ctx, userID uuid.UUID) (*model.UsageSummary, error)
}

type userService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewUserService(db *gorm.DB, validate *validator.Validate) UserService {
	return &userService{
		Log:      utils.Log,
		DB:       db,
		Validate: validate,
	}
}

func (s *userService) GetUsers(c *fiber.Ctx, params *validation.QueryUser) ([]model.User, int64, error) {
	var users []model.User
	var totalResults int64

	if err := s.Validate.Struct(params); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query := s.DB.WithContext(c.Context()).Order("created_at asc")

	if search := params.Search; search != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR role LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	result := query.Find(&users).Count(&totalResults)
	if result.Error != nil {
		s.Log.Errorf("Failed to search users: %+v", result.Error)
		return nil, 0, result.Error
	}

	result = query.Limit(params.Limit).Offset(offset).Find(&users)
	if result.Error != nil {
		s.Log.Errorf("Failed to get all users: %+v", result.Error)
		return nil, 0, result.Error
	}

	return users, totalResults, result.Error
}

func (s *userService) GetUserByID(c *fiber.Ctx, id string) (*model.User, error) {
	user := new(model.User)

	result := s.DB.WithContext(c.Context()).First(user, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if result.Error != nil {
		s.Log.Errorf("Failed get user by id: %+v", result.Error)
	}

	return user, result.Error
}

func (s *userService) GetUserByEmail(c *fiber.Ctx, email string) (*model.User, error) {
	user := new(model.User)

	result := s.DB.WithContext(c.Context()).Where("email = ?", email).First(user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if result.Error != nil {
		s.Log.Errorf("Failed get user by email: %+v", result.Error)
	}

	return user, result.Error
}

func (s *userService) CreateUser(c *fiber.Ctx, req *validation.CreateUser) (*model.User, error) {
	if err := s.Validate.Struct(req); err != nil {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.Log.Errorf("Failed hash password: %+v", err)
		return nil, err
	}

	user := &model.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      req.Role,
	}

	result := s.DB.WithContext(c.Context()).Create(user)

	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return nil, fiber.NewError(fiber.StatusConflict, "Email is already in use")
	}

	if result.Error != nil {
		s.Log.Errorf("Failed to create user: %+v", result.Error)
	}

	return user, result.Error
}

func (s *userService) UpdateUser(c *fiber.Ctx, req *validation.UpdateUser, id string) (*model.User, error) {
	if err := s.Validate.Struct(req); err != nil {
		return nil, err
	}

	if req.Email == "" && req.FirstName == "" && req.Password == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid Request")
	}

	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		req.Password = hashedPassword
	}

	updateBody := &model.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
		Email:     req.Email,
	}

	result := s.DB.WithContext(c.Context()).Where("id = ?", id).Updates(updateBody)

	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return nil, fiber.NewError(fiber.StatusConflict, "Email is already in use")
	}

	if result.RowsAffected == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if result.Error != nil {
		s.Log.Errorf("Failed to update user: %+v", result.Error)
	}

	user, err := s.GetUserByID(c, id)
	if err != nil {
		return nil, err
	}

	return user, result.Error
}

func (s *userService) UpdatePassOrVerify(c *fiber.Ctx, req *validation.UpdatePassOrVerify, id string) error {
	if err := s.Validate.Struct(req); err != nil {
		return err
	}

	if req.Password == "" && !req.EmailVerified {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid Request")
	}

	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return err
		}
		req.Password = hashedPassword
	}

	updateBody := &model.User{
		Password:      req.Password,
		EmailVerified: req.EmailVerified,
	}

	result := s.DB.WithContext(c.Context()).Where("id = ?", id).Updates(updateBody)

	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if result.Error != nil {
		s.Log.Errorf("Failed to update user password or verifiedEmail: %+v", result.Error)
	}

	return result.Error
}

func (s *userService) DeleteUser(c *fiber.Ctx, id string) error {
	user := new(model.User)

	result := s.DB.WithContext(c.Context()).Delete(user, "id = ?", id)

	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if result.Error != nil {
		s.Log.Errorf("Failed to delete user: %+v", result.Error)
	}

	return result.Error
}

func (s *userService) CreateGoogleUser(c *fiber.Ctx, req *validation.GoogleLogin) (*model.User, error) {
	if err := s.Validate.Struct(req); err != nil {
		return nil, err
	}

	userFromDB, err := s.GetUserByEmail(c, req.Email)
	if err != nil {
		if err.Error() == "User not found" {
			user := &model.User{
				FirstName:     req.FirstName,
				LastName:      req.LastName,
				Email:         req.Email,
				EmailVerified: req.EmailVerified,
			}

			if createErr := s.DB.WithContext(c.Context()).Create(user).Error; createErr != nil {
				s.Log.Errorf("Failed to create user: %+v", createErr)
				return nil, createErr
			}

			return user, nil
		}

		return nil, err
	}

	userFromDB.EmailVerified = req.EmailVerified
	if updateErr := s.DB.WithContext(c.Context()).Save(userFromDB).Error; updateErr != nil {
		s.Log.Errorf("Failed to update user: %+v", updateErr)
		return nil, updateErr
	}

	return userFromDB, nil
}

func (t *userService) GetUsageHistory(c *fiber.Ctx, userID uuid.UUID) (*model.UsageSummary, error) {
	// Get total operations and points from existing tasks
	var totalOperations int64
	if err := t.DB.WithContext(c.Context()).Model(&model.Task{}).
		Where("user_id = ?", userID).
		Count(&totalOperations).Error; err != nil {
		return nil, err
	}

	// Calculate total points based on task types
	var tasks []model.Task
	if err := t.DB.WithContext(c.Context()).Model(&model.Task{}).
		Where("user_id = ?", userID).
		Select("type").
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	totalPoints := 0
	for _, task := range tasks {
		totalPoints += model.GetPointsForTaskType(task.Type)
	}

	// Get usage by tool from existing tasks
	var toolUsage []struct {
		TaskType string `json:"taskType"`
		Count    int64  `json:"count"`
	}

	if err := t.DB.WithContext(c.Context()).Model(&model.Task{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Order("count DESC").
		Scan(&toolUsage).Error; err != nil {
		return nil, err
	}

	// Convert to ToolUsage format with display names
	usageByTool := make([]model.ToolUsage, len(toolUsage))
	var mostUsedTool string
	var maxUsage int64

	for i, usage := range toolUsage {
		displayName := model.GetDisplayNameForTaskType(model.TaskType(usage.TaskType))
		percentage := 0.0
		if totalOperations > 0 {
			percentage = float64(usage.Count) / float64(totalOperations) * 100
		}

		usageByTool[i] = model.ToolUsage{
			ToolName:   displayName,
			UsageCount: int(usage.Count),
			Percentage: percentage,
		}

		if usage.Count > maxUsage {
			maxUsage = usage.Count
			mostUsedTool = displayName
		}
	}

	// Get recent activity (last 7 days) from existing tasks
	var recentActivity []struct {
		Date       string `json:"date"`
		Operations int64  `json:"operations"`
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := t.DB.WithContext(c.Context()).Model(&model.Task{}).
		Select("DATE(created_at) as date, COUNT(*) as operations").
		Where("user_id = ? AND created_at >= ?", userID, sevenDaysAgo).
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&recentActivity).Error; err != nil {
		return nil, err
	}

	// Convert to DailyActivity format
	dailyActivity := make([]model.DailyActivity, len(recentActivity))
	for i, activity := range recentActivity {
		dailyActivity[i] = model.DailyActivity{
			Date:       activity.Date,
			Operations: int(activity.Operations),
		}
	}

	return &model.UsageSummary{
		TotalOperations: int(totalOperations),
		PointsUsed:      totalPoints,
		MostUsedTool:    mostUsedTool,
		UsageByTool:     usageByTool,
		RecentActivity:  dailyActivity,
	}, nil
}
