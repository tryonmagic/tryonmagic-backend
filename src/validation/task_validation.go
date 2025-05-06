package validation

import "github.com/mbbipul/prodigyai-backend/src/model"

type ImageSourceType string
type BackgroundSourceType string
type InputData any

const (
	ImageSourceTypeURL    ImageSourceType = "URL"
	ImageSourceTypeBase64 ImageSourceType = "BASE64"
)

const (
	BackgroundSourceTypeImage  BackgroundSourceType = "IMAGE"
	BackgroundSourceTypeColor  BackgroundSourceType = "COLOR"
	BackgroundSourceTypePrompt BackgroundSourceType = "PROMPT"
)

type RemoveBackgroundTaskInput struct {
	SourceImage           string          `json:"sourceImage" validate:"required" example:"https://example.com/image.jpg"`
	ImageSourceType       ImageSourceType `json:"imageSourceType" validate:"required,oneof=URL BASE64" example:"URL"`
	EdgeThreshold         int             `json:"edgeThreshold" validate:"required,min=0,max=100" example:"50"`
	DetailPreservation    bool            `json:"detailPreservation" validate:"required" example:"true"`
	EnhancedHairDetection bool            `json:"enhancedHairDetection" validate:"required" example:"true"`
}

type ReplaceBackgroundTaskInput struct {
	SourceImageUrl            string               `json:"sourceImageUrl" validate:"required" example:"https://example.com/background.jpg"`
	BackgroundSourceType      BackgroundSourceType `json:"backgroundSourceType" validate:"required,oneof=IMAGE COLOR PROMPT" example:"URL"`
	BackgroundImageSourceType *ImageSourceType     `json:"backgroundImageSourceType" validate:"required,oneof=URL BASE64" example:"URL"`
	BackgroundSourceValue     string               `json:"backgroundSourceValue" validate:"required" example:"https://example.com/background.jpg"`
}

type GenerateAIModelTaskInput struct {
	model.AIModel
}

type VTONModelMode string

const (
	ModelModePerformance VTONModelMode = "performance"
	ModelModeQuality     VTONModelMode = "quality"
	ModelModeBalanced    VTONModelMode = "balanced"
)

type VTONTaskInput struct {
	OutfitType   model.OutfitType `json:"outfitType" validate:"required,oneof=TOP BOTTOM FULL_BODY" example:"TOP"`
	ModelMode    VTONModelMode    `json:"modelMode" validate:"required,oneof=performance quality balanced" example:"performance"`
	ModelImage   string           `json:"modelImage" validate:"required" example:"https://example.com/model.jpg"`
	GarmentImage string           `json:"garmentImage" validate:"required" example:"https://example.com/garment.jpg"`
	NumSamples   int              `json:"numSamples" validate:"required,min=1,max=10" example:"5"`
	Seed         int              `json:"seed" validate:"required" example:"42"`
}

type NewTask[T any] struct {
	Type      model.TaskType `json:"type" validate:"required,oneof=REMOVE_BACKGROUND REPLACE_BACKGROUND GENERATE_AI_MODEL VTON" example:"REMOVE_BACKGROUND"`
	InputData T              `json:"inputData" validate:"required" example:"{\"key\": \"value\"}"`
}
