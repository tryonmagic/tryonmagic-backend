package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutfitType string

const (
	OutfitTypeTop      OutfitType = "TOP"
	OutfitTypeBottom   OutfitType = "BOTTOM"
	OutfitTypeFullBody OutfitType = "FULL_BODY"
)

type Garment struct {
	ID         uuid.UUID  `gorm:"primaryKey;not null" json:"id"`
	OutfitType OutfitType `gorm:"not null" json:"outfitType"`
	Color      string     `gorm:"not null" json:"color"`
	ImageUrl   string     `gorm:"not null" json:"imageUrl"`
	Tags       string     `gorm:"not null" json:"tags"`
	UserID     *uuid.UUID `gorm:"not null" json:"userId"`
	User       *User      `gorm:"foreignKey:UserID;references:ID" json:"user"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (aiModel *Garment) BeforeCreate(_ *gorm.DB) error {
	aiModel.ID = uuid.New()
	return nil
}
