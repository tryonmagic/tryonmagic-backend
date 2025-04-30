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

type NewTask[T any] struct {
	Type      model.TaskType `json:"type" validate:"required,oneof=REMOVE_BACKGROUND REPLACE_BACKGROUND GENERATE_AI_MODEL" example:"REMOVE_BACKGROUND"`
	InputData T              `json:"inputData" validate:"required" example:"{\"key\": \"value\"}"`
}
