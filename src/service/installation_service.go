package service

import (
	"fmt"
	"time"

	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AppInstallationService interface defines all app installation operations
type AppInstallationService interface {
	CheckInstallation(shopDomain string) (*validation.InstallationCheckResponse, error)
	CreateInstallation(req *validation.CreateInstallationRequest) (*validation.InstallationResponse, error)
	UpdateInstallation(shopDomain string, req *validation.UpdateInstallationRequest) (*validation.InstallationResponse, error)
	DeleteInstallation(shopDomain string) (*validation.InstallationResponse, error)
	GetInstallation(shopDomain string) (*model.AppInstallation, error)
}

// appInstallationService implements AppInstallationService interface
type appInstallationService struct {
	Log *logrus.Logger
	DB  *gorm.DB
}

// NewAppInstallationService creates a new app installation service instance
func NewAppInstallationService(db *gorm.DB) AppInstallationService {
	return &appInstallationService{
		Log: utils.Log,
		DB:  db,
	}
}

// CheckInstallation checks if a shop has the app installed
func (a *appInstallationService) CheckInstallation(shopDomain string) (*validation.InstallationCheckResponse, error) {
	var installation model.AppInstallation

	err := a.DB.Where("shop = ? AND is_active = ?", shopDomain, true).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &validation.InstallationCheckResponse{
				Installed: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to check installation: %w", err)
	}

	return &validation.InstallationCheckResponse{
		Installed: true,
	}, nil
}

// CreateInstallation creates a new app installation record
func (a *appInstallationService) CreateInstallation(req *validation.CreateInstallationRequest) (*validation.InstallationResponse, error) {
	// Check if installation already exists
	var existingInstallation model.AppInstallation
	err := a.DB.Where("shop = ?", req.Shop).First(&existingInstallation).Error
	if err == nil {
		// Installation exists, update it
		existingInstallation.AccessToken = req.AccessToken
		existingInstallation.Scope = req.Scope
		existingInstallation.IsActive = true
		existingInstallation.UninstalledAt = nil

		if err := a.DB.Save(&existingInstallation).Error; err != nil {
			return nil, fmt.Errorf("failed to update existing installation: %w", err)
		}

		return &validation.InstallationResponse{
			Success: true,
			ID:      existingInstallation.ID.String(),
		}, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing installation: %w", err)
	}

	// Create new installation
	installation := &model.AppInstallation{
		Shop:        req.Shop,
		AccessToken: req.AccessToken,
		Scope:       req.Scope,
		IsActive:    true,
	}

	if err := a.DB.Create(installation).Error; err != nil {
		return nil, fmt.Errorf("failed to create installation: %w", err)
	}

	a.Log.Infof("Created new app installation for shop: %s", req.Shop)

	return &validation.InstallationResponse{
		Success: true,
		ID:      installation.ID.String(),
	}, nil
}

// UpdateInstallation updates an existing app installation
func (a *appInstallationService) UpdateInstallation(shopDomain string, req *validation.UpdateInstallationRequest) (*validation.InstallationResponse, error) {
	var installation model.AppInstallation

	err := a.DB.Where("shop = ?", shopDomain).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("installation not found for shop: %s", shopDomain)
		}
		return nil, fmt.Errorf("failed to find installation: %w", err)
	}

	// Update fields if provided
	if req.AccessToken != nil {
		installation.AccessToken = *req.AccessToken
	}
	if req.Scope != nil {
		installation.Scope = *req.Scope
	}
	if req.IsActive != nil {
		installation.IsActive = *req.IsActive
		if !*req.IsActive {
			now := time.Now()
			installation.UninstalledAt = &now
		} else {
			installation.UninstalledAt = nil
		}
	}

	if err := a.DB.Save(&installation).Error; err != nil {
		return nil, fmt.Errorf("failed to update installation: %w", err)
	}

	a.Log.Infof("Updated app installation for shop: %s", shopDomain)

	return &validation.InstallationResponse{
		Success: true,
	}, nil
}

// DeleteInstallation deletes app installation for a shop (called on app/uninstalled webhook)
func (a *appInstallationService) DeleteInstallation(shopDomain string) (*validation.InstallationResponse, error) {
	// Start a transaction to ensure all deletions succeed or fail together
	tx := a.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var sessionsDeleted int

	// Delete Shopify sessions - database cascade delete handles related records automatically

	result := tx.Where("shop = ?", shopDomain).Delete(&model.ShopifySession{})
	if result.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete Shopify sessions: %w", result.Error)
	}
	sessionsDeleted = int(result.RowsAffected)

	// Mark installation as inactive instead of deleting (for audit trail)
	var installation model.AppInstallation
	err := tx.Where("shop = ?", shopDomain).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Installation not found, but sessions were deleted
			tx.Commit()
			return &validation.InstallationResponse{
				Success:         true,
				SessionsDeleted: sessionsDeleted,
			}, nil
		}
		tx.Rollback()
		return nil, fmt.Errorf("failed to find installation: %w", err)
	}

	// Mark as inactive and set uninstalled timestamp
	now := time.Now()
	installation.IsActive = false
	installation.UninstalledAt = &now

	if err := tx.Save(&installation).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to mark installation as inactive: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	a.Log.Infof("Deleted app installation for shop: %s, sessions deleted: %d", shopDomain, sessionsDeleted)

	return &validation.InstallationResponse{
		Success:         true,
		SessionsDeleted: sessionsDeleted,
	}, nil
}

// GetInstallation retrieves an app installation by shop domain
func (a *appInstallationService) GetInstallation(shopDomain string) (*model.AppInstallation, error) {
	var installation model.AppInstallation

	err := a.DB.Where("shop = ?", shopDomain).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("installation not found for shop: %s", shopDomain)
		}
		return nil, fmt.Errorf("failed to get installation: %w", err)
	}

	return &installation, nil
}
