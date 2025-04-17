package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID           uuid.UUID `gorm:"primaryKey;not null" json:"id"`
	Key          string    `gorm:"unique;not null" json:"key"`
	FullPath     string    `gorm:"not null" json:"fullPath"`
	URL          string    `gorm:"not null" json:"url"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
	ResizeURL    *string   `json:"resizeUrl"`
	Name         string    `gorm:"not null" json:"name"`
	Type         string    `gorm:"not null" json:"type"`
	Size         int       `gorm:"not null" json:"size"`
	Width        int       `gorm:"not null" json:"width"`
	Height       int       `gorm:"not null" json:"height"`
	Ext          string    `gorm:"not null" json:"ext"`
	Mime         string    `gorm:"not null" json:"mime"`
	ThumbHash    *[]byte   `json:"thumbHash"`
	UserID       uuid.UUID `gorm:"not null" json:"userId"`
	User         *User     `gorm:"foreignKey:UserID;references:ID" json:"user"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (user *File) BeforeCreate(_ *gorm.DB) error {
	user.ID = uuid.New() // Generate UUID before create
	return nil
}
