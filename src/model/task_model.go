package model

import (
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

	TaskTypeGenerateAIModel TaskType = "GENERATE_AI_MODEL"
	TaskTypeVTon            TaskType = "VTON"
	TaskTypeUpscale         TaskType = "UPSCALE"
)

type Task struct {
	ID          uuid.UUID       `gorm:"primaryKey;type:char(36)" json:"id"` // UUID
	Type        TaskType        `gorm:"not null" json:"type"`
	Status      TaskStatus      `gorm:"default:PENDING;not null" json:"status"`
	InputData   datatypes.JSON  `gorm:"type:text" json:"inputData"`
	ResultData  *datatypes.JSON `gorm:"type:text" json:"resultData"`
	Error       *string         `gorm:"type:text" json:"error"`
	RetryCount  int             `gorm:"default:0" json:"retryCount"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	StartedAt   *time.Time      `gorm:"index" json:"startedAt"`
	CompletedAt *time.Time      `gorm:"index" json:"completedAt"`
	UserID      uuid.UUID       `gorm:"not null" json:"userId"`
	User        *User           `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (task *Task) BeforeCreate(tx *gorm.DB) error {
	task.ID = uuid.New()
	return nil
}
