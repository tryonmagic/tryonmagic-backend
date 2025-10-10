package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "PENDING"
	TaskStatusQueued     TaskStatus = "QUEUED"
	TaskStatusProcessing TaskStatus = "PROCESSING"
	TaskStatusSucceeded  TaskStatus = "SUCCEEDED"
	TaskStatusFailed     TaskStatus = "FAILED"
)

type TaskType string

const (
	TaskTypeRemoveBackground  TaskType = "REMOVE_BACKGROUND"
	TaskTypeGenerateImage     TaskType = "GENERATE_IMAGE"
	TaskTypeReplaceBackground TaskType = "REPLACE_BACKGROUND"

	TaskTypeGenerateAIModel  TaskType = "GENERATE_AI_MODEL"
	TaskTypeVTon             TaskType = "VTON"
	TaskTypeUpscale          TaskType = "UPSCALE"
	TaskTypeEnhance          TaskType = "ENHANCE"
	TaskTypeExpandImage      TaskType = "EXPAND_IMAGE"
	TaskTypeEditImage        TaskType = "EDIT_IMAGE"
	TaskTypeProductImageEdit TaskType = "PRODUCT_IMAGE_EDITING"
)

type Task struct {
	ID                 uuid.UUID           `gorm:"primaryKey;type:char(36)" json:"id"` // UUID
	Type               TaskType            `gorm:"not null" json:"type"`
	Status             TaskStatus          `gorm:"default:PENDING;not null" json:"status"`
	InputData          datatypes.JSON      `gorm:"type:text" json:"inputData"`
	ResultData         *datatypes.JSON     `gorm:"type:text" json:"resultData"`
	Error              *string             `gorm:"type:text" json:"error"`
	RetryCount         int                 `gorm:"default:0" json:"retryCount"`
	CreatedAt          time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
	StartedAt          *time.Time          `gorm:"index" json:"startedAt"`
	CompletedAt        *time.Time          `gorm:"index" json:"completedAt"`
	UserID             uuid.UUID           `gorm:"not null" json:"userId"`
	User               *User               `gorm:"foreignKey:UserID;references:ID" json:"user"`
	TransactionHistory *TransactionHistory `gorm:"foreignKey:TaskID;references:ID" json:"transactionHistory"`
}

type UserSubscription struct {
	Plan          SubscriptionPlan   `gorm:"not null;default:FREE" json:"plan"`
	Status        SubscriptionStatus `gorm:"not null;default:ACTIVE" json:"status"`
	PointsBalance int                `gorm:"not null;default:0" json:"pointsBalance"`
	PointsUsed    int                `gorm:"not null;default:0" json:"pointsUsed"`
	PointsTotal   int                `gorm:"not null;default:0" json:"pointsTotal"`
	MonthlyPoints int                `gorm:"not null;default:0" json:"monthlyPoints"`
	LastResetDate time.Time          `gorm:"not null" json:"lastResetDate"`
	NextResetDate time.Time          `gorm:"not null" json:"nextResetDate"`
}

// UsageSummary represents aggregated usage data for dashboard
type UsageSummary struct {
	TotalPoints         int              `json:"totalPoints"`
	CurrentSubscription UserSubscription `json:"currentSubscription"`
	TotalOperations     int              `json:"totalOperations"`
	PointsUsed          int              `json:"pointsUsed"`
	MostUsedTool        ToolUsage        `json:"mostUsedTool"`
	UsageByTool         []ToolUsage      `json:"usageByTool"`
	DailyActivity       []DailyActivity  `json:"dailyActivity"`
	RecentOperations    []Task           `json:"recentOperations"`
}

// ToolUsage represents usage statistics for a specific tool
type ToolUsage struct {
	ToolName   string  `json:"toolName"`
	UsageCount int     `json:"usageCount"`
	Percentage float64 `json:"percentage"`
}

// DailyActivity represents daily usage activity
type DailyActivity struct {
	Date       string `json:"date"`
	Operations int    `json:"operations"`
}

func (task *Task) BeforeCreate(tx *gorm.DB) error {
	task.ID = uuid.New()
	return nil
}

func (task *Task) GetPointsForTaskType() int {
	return GetPointsForTaskType(task.Type)
}

func GetPointsForTaskType(taskType TaskType) int {
	switch taskType {
	case TaskTypeRemoveBackground:
		return 1
	case TaskTypeGenerateAIModel:
		return 2
	case TaskTypeVTon:
		return 3
	case TaskTypeUpscale:
		return 1
	case TaskTypeEnhance:
		return 2
	case TaskTypeExpandImage:
		return 2
	case TaskTypeEditImage:
		return 2
	case TaskTypeProductImageEdit:
		return 3
	case TaskTypeReplaceBackground:
		return 1
	case TaskTypeGenerateImage:
		return 2
	default:
		return 1
	}
}

func (task *Task) GetDisplayNameForTaskType() string {
	return GetDisplayNameForTaskType(task.Type)
}

func GetDisplayNameForTaskType(taskType TaskType) string {
	fmt.Println("taskType", taskType)
	switch taskType {
	case TaskTypeRemoveBackground:
		return "Remove Background"
	case TaskTypeGenerateAIModel:
		return "Generate AI Model"
	case TaskTypeVTon:
		return "V-Toning"
	case TaskTypeUpscale:
		return "Upscale"
	case TaskTypeEnhance:
		return "Enhance"
	case TaskTypeExpandImage:
		return "Expand Image"
	case TaskTypeEditImage:
		return "Edit Image"
	case TaskTypeProductImageEdit:
		return "Product Image Editing"
	case TaskTypeReplaceBackground:
		return "Replace Background"
	case TaskTypeGenerateImage:
		return "Generate Image"
	default:
		return "Unknown"
	}
}
