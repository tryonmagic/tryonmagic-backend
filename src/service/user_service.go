package service

import (
	"context"
	"errors"
	"fmt"
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
	GetTasks(c *fiber.Ctx, userID uuid.UUID, queryParams validation.GetTasksQueryParams) (*PaginationResult[model.Task], error)
	AuditTask(c *fiber.Ctx, userID uuid.UUID, taskID uuid.UUID, req *validation.AuditTaskRequest) error
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
	query := s.DB.WithContext(context.Background()).Order("created_at asc")

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

	result := s.DB.WithContext(context.Background()).Preload("Subscription").First(user, "id = ?", id)

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

	result := s.DB.WithContext(context.Background()).Where("email = ?", email).First(user)

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

	result := s.DB.WithContext(context.Background()).Create(user)

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

	result := s.DB.WithContext(context.Background()).Where("id = ?", id).Updates(updateBody)

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

	result := s.DB.WithContext(context.Background()).Where("id = ?", id).Updates(updateBody)

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

	result := s.DB.WithContext(context.Background()).Delete(user, "id = ?", id)

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

			if createErr := s.DB.WithContext(context.Background()).Create(user).Error; createErr != nil {
				s.Log.Errorf("Failed to create user: %+v", createErr)
				return nil, createErr
			}

			return user, nil
		}

		return nil, err
	}

	userFromDB.EmailVerified = req.EmailVerified
	if updateErr := s.DB.WithContext(context.Background()).Save(userFromDB).Error; updateErr != nil {
		s.Log.Errorf("Failed to update user: %+v", updateErr)
		return nil, updateErr
	}

	return userFromDB, nil
}

func (t *userService) GetUsageHistory(c *fiber.Ctx, userID uuid.UUID) (*model.UsageSummary, error) {
	// Get total operations and points from existing tasks
	var totalOperations int64
	if err := t.DB.WithContext(context.Background()).Model(&model.Task{}).
		Where("user_id = ?", userID).
		Count(&totalOperations).Error; err != nil {
		return nil, err
	}

	var transactionHistory []model.TransactionHistory
	if err := t.DB.WithContext(context.Background()).Model(&model.TransactionHistory{}).
		Select("task_id, transaction_type, transaction_status, net_amount").
		Where("user_id = ?", userID).
		Find(&transactionHistory).Error; err != nil {
		t.Log.Errorf("Failed to get transaction history: %+v", err)
		return nil, err
	}

	var userSubscription model.UserSubscription
	if err := t.DB.WithContext(context.Background()).Model(&model.Subscription{}).
		Select("points_balance, points_used,status,plan, points_total, monthly_points, last_reset_date, next_reset_date, expires_at").
		Where("user_id = ?", userID).
		First(&userSubscription).Error; err != nil {
		t.Log.Errorf("Failed to get user subscription: %+v", err)
		return nil, err
	}

	totalCreditPoints := 0
	for _, transaction := range transactionHistory {
		if transaction.TransactionType == model.TransactionTypeTaskExecution {
			totalCreditPoints += -1 * transaction.NetAmount
		}
	}

	// Get usage by tool from existing tasks
	var toolUsage []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	if err := t.DB.WithContext(context.Background()).Model(&model.Task{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Order("count DESC").
		Scan(&toolUsage).Error; err != nil {
		return nil, err
	}

	// Convert to ToolUsage format with display names
	usageByTool := make([]model.ToolUsage, len(toolUsage))
	var mostUsedTool model.ToolUsage
	var maxUsage int64

	for i, usage := range toolUsage {
		displayName := model.GetDisplayNameForTaskType(model.TaskType(usage.Type))
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
			mostUsedTool = model.ToolUsage{
				ToolName:   displayName,
				UsageCount: int(usage.Count),
				Percentage: percentage,
			}
		}
	}

	var recentTasks []model.Task
	if err := t.DB.WithContext(context.Background()).Model(&model.Task{}).
		Preload("TransactionHistory").
		Where("user_id = ?", userID).
		Order("tasks.created_at DESC").
		Limit(10).
		Find(&recentTasks).Error; err != nil {
		return nil, err
	}

	// Get recent activity (last 7 days) from existing tasks
	var dailyTask []struct {
		Date       string `json:"date"`
		Operations int64  `json:"operations"`
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := t.DB.WithContext(context.Background()).Model(&model.Task{}).
		Select("DATE(created_at) as date, COUNT(*) as operations").
		Where("user_id = ? AND created_at >= ?", userID, sevenDaysAgo).
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&dailyTask).Error; err != nil {
		return nil, err
	}

	// Convert to DailyActivity format
	dailyActivity := make([]model.DailyActivity, len(dailyTask))
	for i, activity := range dailyTask {
		dailyActivity[i] = model.DailyActivity{
			Date:       activity.Date,
			Operations: int(activity.Operations),
		}
	}
	fmt.Printf("dailyActivity: %+v\n", dailyActivity)

	return &model.UsageSummary{
		TotalOperations:     int(totalOperations),
		PointsUsed:          totalCreditPoints,
		MostUsedTool:        mostUsedTool,
		UsageByTool:         usageByTool,
		DailyActivity:       dailyActivity,
		RecentOperations:    recentTasks,
		CurrentSubscription: userSubscription,
	}, nil
}

func (t *userService) GetTasks(c *fiber.Ctx, userID uuid.UUID, queryParams validation.GetTasksQueryParams) (*PaginationResult[model.Task], error) {
	var tasks []model.Task
	fmt.Printf("userID: %+v\n", userID)
	fmt.Printf("queryParams: %+v\n", queryParams)
	query := t.DB.WithContext(context.Background()).Model(&model.Task{}).Where("user_id = ?", userID)

	if queryParams.TaskStatus != "" {
		query = query.Where("status = ?", queryParams.TaskStatus)
	}

	if queryParams.TaskType != "" {
		query = query.Where("type = ?", queryParams.TaskType)
	}

	if queryParams.DateFrom != nil && *queryParams.DateFrom != "" {
		query = query.Where("created_at >= ?", *queryParams.DateFrom)
	}

	if queryParams.DateTo != nil && *queryParams.DateTo != "" {
		query = query.Where("created_at <= ?", *queryParams.DateTo)
	}

	query = query.Order("tasks.created_at DESC")

	var count int64
	if err := query.Count(&count).Error; err != nil {
		t.Log.Errorf("Failed to get tasks count: %+v", err)
		return nil, err
	}
	offset := GetPaginationOffset(queryParams.Page, queryParams.PageSize)
	query = query.Offset(offset).Limit(queryParams.PageSize)
	if err := query.Preload("TransactionHistory").Preload("TaskAudit").Find(&tasks).Error; err != nil {
		t.Log.Errorf("Failed to get tasks: %+v", err)
		return nil, err
	}
	return &PaginationResult[model.Task]{
		Results: tasks,
		Count:   count,
	}, nil
}

func (s *userService) AuditTask(c *fiber.Ctx, userID uuid.UUID, taskID uuid.UUID, req *validation.AuditTaskRequest) error {
	if err := s.Validate.Struct(req); err != nil {
		return err
	}

	// 1. Verify task exists and belongs to the user
	var task model.Task
	if err := s.DB.WithContext(context.Background()).Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Task not found")
		}
		return err
	}

	// 2. Create TaskAudit record
	audit := &model.TaskAudit{
		TaskID:            taskID,
		UserID:            userID,
		Reason:            req.Reason,
		IsRefundRequested: req.IsRefundRequested,
		Status:            model.AuditStatusPending,
	}

	if err := s.DB.WithContext(context.Background()).Create(audit).Error; err != nil {
		s.Log.Errorf("Failed to create task audit: %+v", err)
		return err
	}

	// 3. Handle refund if requested (optional: auto-refund can be added here based on business rules)
	// For now, we just log it and store it in the audit table.
	// If the user specifically wants an auto-refund for certain conditions, we could add that logic.

	return nil
}
