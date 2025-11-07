package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GDPRService interface defines all GDPR-related operations
type GDPRService interface {
	LogGDPRRequest(req *validation.LogGDPRRequest) (*validation.GDPRRequestResponse, error)
	UpdateRequestStatus(requestID string, req *validation.UpdateGDPRRequestStatus) (*validation.GDPRStatusResponse, error)
	GetCustomerData(customerID, shop string) (*validation.CustomerDataResponse, error)
	RedactCustomerData(customerID string, req *validation.RedactCustomerDataRequest) (*validation.RedactionResponse, error)
	RedactShopData(shop string) (*validation.RedactionResponse, error)
	CleanupOldLogs() (*validation.CleanupResponse, error)
}

// gdprService implements GDPRService interface
type gdprService struct {
	Log *logrus.Logger
	DB  *gorm.DB
}

// NewGDPRService creates a new GDPR service instance
func NewGDPRService(db *gorm.DB) GDPRService {
	return &gdprService{
		Log: utils.Log,
		DB:  db,
	}
}

// LogGDPRRequest creates a new GDPR request log entry
func (g *gdprService) LogGDPRRequest(req *validation.LogGDPRRequest) (*validation.GDPRRequestResponse, error) {
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	gdprRequest := &model.GDPRRequest{
		Shop:          req.Shop,
		CustomerID:    req.CustomerID,
		CustomerEmail: req.CustomerEmail,
		RequestType:   req.RequestType,
		Status:        model.GDPRRequestStatusPending,
		Payload:       datatypes.JSON(payloadBytes),
	}

	if err := g.DB.Create(gdprRequest).Error; err != nil {
		return nil, fmt.Errorf("failed to create GDPR request: %w", err)
	}

	return &validation.GDPRRequestResponse{
		ID: gdprRequest.ID.String(),
	}, nil
}

// UpdateRequestStatus updates the status of a GDPR request
func (g *gdprService) UpdateRequestStatus(requestID string, req *validation.UpdateGDPRRequestStatus) (*validation.GDPRStatusResponse, error) {
	requestUUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, fmt.Errorf("invalid request ID format: %w", err)
	}

	updates := map[string]interface{}{
		"status": req.Status,
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if req.Status == model.GDPRRequestStatusCompleted || req.Status == model.GDPRRequestStatusFailed {
		now := time.Now()
		updates["processed_at"] = &now
	}

	if err := g.DB.Model(&model.GDPRRequest{}).Where("id = ?", requestUUID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update GDPR request status: %w", err)
	}

	return &validation.GDPRStatusResponse{
		Success: true,
	}, nil
}

// GetCustomerData gathers all customer data for a GDPR data request
func (g *gdprService) GetCustomerData(customerID, shop string) (*validation.CustomerDataResponse, error) {
	var preferences []model.UserPreferences
	var images []model.GeneratedImage
	var usage []model.UsageHistory
	var shopSettings model.ShopSettings

	// Get user preferences
	if err := g.DB.Where("shop = ? AND customer_id = ?", shop, customerID).Find(&preferences).Error; err != nil {
		g.Log.Errorf("Failed to fetch user preferences: %v", err)
	}

	// Get generated images
	if err := g.DB.Where("shop = ? AND customer_id = ?", shop, customerID).Find(&images).Error; err != nil {
		g.Log.Errorf("Failed to fetch generated images: %v", err)
	}

	// Get usage history
	if err := g.DB.Where("shop = ? AND customer_id = ?", shop, customerID).Find(&usage).Error; err != nil {
		g.Log.Errorf("Failed to fetch usage history: %v", err)
	}

	// Get shop settings
	if err := g.DB.Where("shop = ?", shop).First(&shopSettings).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			g.Log.Errorf("Failed to fetch shop settings: %v", err)
		}
		// Set to nil if not found
		shopSettings = model.ShopSettings{}
	}

	data := validation.CustomerDataCollection{
		Preferences:  preferences,
		Images:       images,
		Usage:        usage,
		ShopSettings: &shopSettings,
	}

	// Only include shop settings if it has an ID (was found)
	if shopSettings.ID == uuid.Nil {
		data.ShopSettings = nil
	}

	return &validation.CustomerDataResponse{
		Shop:        shop,
		CustomerID:  customerID,
		CollectedAt: time.Now().Format(time.RFC3339),
		Data:        data,
	}, nil
}

// RedactCustomerData deletes or anonymizes all customer data for GDPR compliance
func (g *gdprService) RedactCustomerData(customerID string, req *validation.RedactCustomerDataRequest) (*validation.RedactionResponse, error) {
	// Start a transaction to ensure all deletions succeed or fail together
	tx := g.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete user preferences
	if err := tx.Where("shop = ? AND customer_id = ?", req.Shop, customerID).Delete(&model.UserPreferences{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete user preferences: %w", err)
	}

	// Delete generated images
	if err := tx.Where("shop = ? AND customer_id = ?", req.Shop, customerID).Delete(&model.GeneratedImage{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete generated images: %w", err)
	}

	// Delete usage history
	if err := tx.Where("shop = ? AND customer_id = ?", req.Shop, customerID).Delete(&model.UsageHistory{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete usage history: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	g.Log.Infof("Successfully redacted customer data for customer %s in shop %s", customerID, req.Shop)

	return &validation.RedactionResponse{
		Success: true,
	}, nil
}

// RedactShopData deletes all shop-related data
func (g *gdprService) RedactShopData(shop string) (*validation.RedactionResponse, error) {
	// Start a transaction to ensure all deletions succeed or fail together
	tx := g.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var deletedCount int

	// Delete app installation
	if err := tx.Where("shop = ?", shop).Delete(&model.AppInstallation{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete app installation: %w", err)
	}

	// Delete shop settings
	if err := tx.Where("shop = ?", shop).Delete(&model.ShopSettings{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete shop settings: %w", err)
	}

	// Count and delete user preferences
	result := tx.Where("shop = ?", shop).Delete(&model.UserPreferences{})
	if result.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete user preferences: %w", result.Error)
	}
	deletedCount += int(result.RowsAffected)

	// Count and delete generated images
	result = tx.Where("shop = ?", shop).Delete(&model.GeneratedImage{})
	if result.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete generated images: %w", result.Error)
	}
	deletedCount += int(result.RowsAffected)

	// Count and delete usage history
	result = tx.Where("shop = ?", shop).Delete(&model.UsageHistory{})
	if result.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete usage history: %w", result.Error)
	}
	deletedCount += int(result.RowsAffected)

	// Count and delete GDPR requests
	result = tx.Where("shop = ?", shop).Delete(&model.GDPRRequest{})
	if result.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete GDPR requests: %w", result.Error)
	}
	deletedCount += int(result.RowsAffected)

	// Log the deletion
	shopDeletion := &model.ShopDeletion{
		Shop:        shop,
		Reason:      "GDPR shop redaction",
		DataDeleted: deletedCount,
	}
	if err := tx.Create(shopDeletion).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to log shop deletion: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	g.Log.Infof("Successfully redacted shop data for shop %s, deleted %d records", shop, deletedCount)

	return &validation.RedactionResponse{
		Success: true,
	}, nil
}

// CleanupOldLogs removes GDPR request logs older than 2 years
func (g *gdprService) CleanupOldLogs() (*validation.CleanupResponse, error) {
	twoYearsAgo := time.Now().AddDate(-2, 0, 0)

	result := g.DB.Where("created_at < ?", twoYearsAgo).Delete(&model.GDPRRequest{})
	if result.Error != nil {
		return nil, fmt.Errorf("failed to cleanup old GDPR logs: %w", result.Error)
	}

	deletedCount := int(result.RowsAffected)
	g.Log.Infof("Cleaned up %d old GDPR request logs", deletedCount)

	return &validation.CleanupResponse{
		Success:      true,
		DeletedCount: deletedCount,
	}, nil
}
