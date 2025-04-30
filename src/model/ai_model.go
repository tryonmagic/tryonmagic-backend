package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AIModelAuthor string

const (
	AIModelAuthorUser   AIModelAuthor = "USER"
	AIModelAuthorSystem AIModelAuthor = "SYSTEM"
)

type AIModel struct {
	ID                 uuid.UUID      `gorm:"primaryKey;not null" json:"id"`
	AIModelAuthor      AIModelAuthor  `gorm:"default:SYSTEM;not null" json:"aiModelAuthor"`
	ModelTitle         string         `gorm:"not null" json:"modelTitle"`
	Prompt             string         `gorm:"not null" json:"prompt"`
	NegativePrompt     *string        `json:"negativePrompt"`
	GenAIName          string         `gorm:"not null" json:"genAIName"`
	ModelCategory      string         `json:"modelCategory"` // "fashion", "lifestyle", "travel"
	ModelLocation      string         `json:"modelLocation"` // (e.g., "studio", "beach", "urban rooftop")
	Ethnicity          string         `json:"ethnicity"`
	SkinTone           *string        `json:"skinTone"`
	Gender             string         `json:"gender"`
	AgeRange           string         `json:"ageRange"`
	IsActive           bool           `gorm:"default:true" json:"isActive"`
	StyleTags          string         `json:"styleTags"`
	OverrideParameters bool           `json:"overrideParameters"`
	Resolution         datatypes.JSON `json:"resolution"`
	AspectRatio        string         `json:"aspectRatio"`
	ModelImageUrl      *string        `json:"modelImageUrl"`
	UserID             *uuid.UUID     `gorm:"not null" json:"userId"`
	TaskID             *uuid.UUID     `json:"taskId"`
	User               *User          `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Task               *Task          `gorm:"foreignKey:TaskID;references:ID" json:"task"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (aiModel *AIModel) BeforeCreate(_ *gorm.DB) error {
	aiModel.ID = uuid.New()
	return nil
}
