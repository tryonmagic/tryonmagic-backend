package validation

import (
	"github.com/mbbipul/tryonmagic-backend/src/model"
)

// GDPR Request Validation Structures

// LogGDPRRequest represents the request body for logging a GDPR request
type LogGDPRRequest struct {
	Shop          string                 `json:"shop" validate:"required"`
	CustomerID    string                 `json:"customerId" validate:"required"`
	CustomerEmail string                 `json:"customerEmail" validate:"required,email"`
	RequestType   model.GDPRRequestType  `json:"requestType" validate:"required,oneof=data_request redact shop_redact"`
	Payload       map[string]interface{} `json:"payload"`
}

// UpdateGDPRRequestStatus represents the request body for updating GDPR request status
type UpdateGDPRRequestStatus struct {
	Status model.GDPRRequestStatus `json:"status" validate:"required,oneof=pending processing completed failed"`
	Notes  *string                 `json:"notes"`
}

// GetCustomerDataQuery represents query parameters for getting customer data
type GetCustomerDataQuery struct {
	Shop string `query:"shop" validate:"required"`
}

// RedactCustomerDataRequest represents the request body for customer data redaction
type RedactCustomerDataRequest struct {
	Shop string `json:"shop" validate:"required"`
}

// App Installation Validation Structures

// CreateInstallationRequest represents the request body for creating an app installation
type CreateInstallationRequest struct {
	Shop        string `json:"shop" validate:"required"`
	AccessToken string `json:"accessToken" validate:"required"`
	Scope       string `json:"scope" validate:"required"`
}

// UpdateInstallationRequest represents the request body for updating an app installation
type UpdateInstallationRequest struct {
	AccessToken *string `json:"accessToken"`
	Scope       *string `json:"scope"`
	IsActive    *bool   `json:"isActive"`
}

// GDPR Response Structures

// GDPRRequestResponse represents the response for GDPR request operations
type GDPRRequestResponse struct {
	ID string `json:"id"`
}

// GDPRStatusResponse represents the response for status update operations
type GDPRStatusResponse struct {
	Success bool `json:"success"`
}

// CustomerDataResponse represents the response for customer data retrieval
type CustomerDataResponse struct {
	Shop        string                 `json:"shop"`
	CustomerID  string                 `json:"customerId"`
	CollectedAt string                 `json:"collectedAt"`
	Data        CustomerDataCollection `json:"data"`
}

// CustomerDataCollection represents the collected customer data
type CustomerDataCollection struct {
	Preferences  []model.UserPreferences `json:"preferences"`
	Images       []model.GeneratedImage  `json:"images"`
	Usage        []model.UsageHistory    `json:"usage"`
	ShopSettings *model.ShopSettings     `json:"shopSettings"`
}

// RedactionResponse represents the response for data redaction operations
type RedactionResponse struct {
	Success bool `json:"success"`
}

// CleanupResponse represents the response for cleanup operations
type CleanupResponse struct {
	Success      bool `json:"success"`
	DeletedCount int  `json:"deletedCount"`
}

// App Installation Response Structures

// InstallationCheckResponse represents the response for installation check
type InstallationCheckResponse struct {
	Installed bool `json:"installed"`
}

// InstallationResponse represents the response for installation operations
type InstallationResponse struct {
	Success         bool   `json:"success"`
	SessionsDeleted int    `json:"sessionsDeleted,omitempty"`
	ID              string `json:"id,omitempty"`
}

// Error Response Structure

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
