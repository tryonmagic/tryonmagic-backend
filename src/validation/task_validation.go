package validation

import "github.com/mbbipul/tryonmagic-backend/src/model"

type ImageSourceType string
type BackgroundSourceType string
type InputData any
type EditMode string
type EditType string

const (
	ImageSourceTypeURL    ImageSourceType = "URL"
	ImageSourceTypeBase64 ImageSourceType = "BASE64"
)

const (
	BackgroundSourceTypeImage  BackgroundSourceType = "IMAGE"
	BackgroundSourceTypeColor  BackgroundSourceType = "COLOR"
	BackgroundSourceTypePrompt BackgroundSourceType = "PROMPT"
)

const (
	EditModePrompt EditMode = "PROMPT"
	EditModeRegion EditMode = "REGION"
)

const (
	EditTypeInpaintRemove  EditType = "EDIT_MODE_INPAINT_REMOVAL"
	EditTypeInpaintReplace EditType = "EDIT_MODE_INPAINT_REPLACE"
	EditTypeInpaintAdd     EditType = "EDIT_MODE_INPAINT_INSERTION"
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

type UpscaleTaskInput struct {
	SourceImageUrl string `json:"sourceImageUrl" validate:"required" example:"https://example.com/image.jpg"`
	// Scale           string          `json:"scale" validate:"required,oneof=2x 4x" example:"2x"`
	ImageSourceType ImageSourceType `json:"imageSourceType" validate:"required,oneof=URL BASE64" example:"URL"`
}

type EditRegion struct {
	X      int `json:"x" validate:"required,min=0" example:"100"`
	Y      int `json:"y" validate:"required,min=0" example:"100"`
	Width  int `json:"width" validate:"required,min=1" example:"200"`
	Height int `json:"height" validate:"required,min=1" example:"200"`
}

type ReplacePrompt struct {
	ReplaceFrom *string `json:"replaceFrom" example:"cat"`
	ReplaceTo   string  `json:"replaceTo" example:"dog"`
}

type EditImageTaskInput struct {
	SourceImageUrl  string          `json:"sourceImageUrl" validate:"required" example:"https://example.com/image.jpg"`
	ImageSourceType ImageSourceType `json:"imageSourceType" validate:"required,oneof=URL BASE64" example:"URL"`
	EditType        EditType        `json:"editType" validate:"required,oneof=EDIT_MODE_INPAINT_REMOVAL EDIT_MODE_INPAINT_REPLACE EDIT_MODE_INPAINT_INSERTION" example:"EDIT_MODE_INPAINT_REMOVAL"`
	EditMode        EditMode        `json:"editMode" validate:"required,oneof=PROMPT REGION" example:"PROMPT"`
	EditPrompt      *string         `json:"editPrompt"  example:"A beautiful landscape"`
	EditRegion      []EditRegion    `json:"editRegion"  example:"[{\"x\": 100, \"y\": 100, \"width\": 200, \"height\": 200}]"`
	ReplacePrompt   *ReplacePrompt  `json:"replacePrompt" example:"{\"replaceFrom\": \"\", \"replaceTo\": \"\"}"`
}

type NewTask[T any] struct {
	Type      model.TaskType `json:"type" validate:"required,oneof=REMOVE_BACKGROUND REPLACE_BACKGROUND GENERATE_AI_MODEL VTON UPSCALE ENHANCE EXPAND_IMAGE EDIT_IMAGE PRODUCT_IMAGE_EDITING" example:"REMOVE_BACKGROUND"`
	InputData T              `json:"inputData" validate:"required" example:"{\"key\": \"value\"}"`
}

type ExpandImageTaskInput struct {
	SourceImageUrl      string          `json:"sourceImageUrl" validate:"required" example:"https://example.com/image.jpg"`
	TargetWidth         int             `json:"targetWidth" validate:"required,min=1" example:"800"`
	TargetHeight        int             `json:"targetHeight" validate:"required,min=1" example:"600"`
	HorizontalAlignment string          `json:"horizontalAlignment" validate:"required,oneof=left center right" example:"center"`
	VerticalAlignment   string          `json:"verticalAlignment" validate:"required,oneof=top center bottom" example:"center"`
	ExpandPrompt        string          `json:"expandPrompt" validate:"required" example:"A beautiful landscape"`
	ImageSourceType     ImageSourceType `json:"imageSourceType" validate:"required,oneof=URL BASE64" example:"URL"`
}

// Product Image Editing Types
type ProductImageEditingType string

const (
	SmartBackground      ProductImageEditingType = "SMART_BACKGROUND"
	StyleTransfer        ProductImageEditingType = "STYLE_TRANSFER"
	SceneGeneration      ProductImageEditingType = "SCENE_GENERATION"
	ProductVariants      ProductImageEditingType = "PRODUCT_VARIANTS"
	ProductColorVariants ProductImageEditingType = "PRODUCT_COLOR_VARIANTS"
)

type StyleType string

const (
	Artistic StyleType = "ARTISTIC"
	Mood     StyleType = "MOOD"
	Brand    StyleType = "BRAND"
	Creative StyleType = "CREATIVE"
)

type SceneType string

const (
	Lifestyle        SceneType = "LIFESTYLE"
	Commercial       SceneType = "COMMERCIAL"
	Context          SceneType = "CONTEXT"
	BrandEnvironment SceneType = "BRAND_ENVIRONMENT"
)

type VariantType string

const (
	Size          VariantType = "SIZE"
	Configuration VariantType = "CONFIGURATION"
	Angle         VariantType = "ANGLE"
	Batch         VariantType = "BATCH"
)

type ProductImageEditTaskInput struct {
	// Common fields
	SourceImageUrl  string                  `json:"sourceImageUrl" validate:"required" example:"https://example.com/image.jpg"`
	ImageSourceType ImageSourceType         `json:"imageSourceType" validate:"required,oneof=URL BASE64" example:"URL"`
	ProcessingType  ProductImageEditingType `json:"processingType" validate:"required,oneof=SMART_BACKGROUND STYLE_TRANSFER SCENE_GENERATION PRODUCT_VARIANTS PRODUCT_COLOR_VARIANTS" example:"SMART_BACKGROUND"`

	// Smart Background fields
	BackgroundPrompt   *string `json:"backgroundPrompt" example:"A modern office background"`
	PreserveLighting   *bool   `json:"preserveLighting" example:"true"`
	MatchShadows       *bool   `json:"matchShadows" example:"true"`
	EnhanceQuality     *bool   `json:"enhanceQuality" example:"true"`
	BackgroundStyle    *string `json:"backgroundStyle" example:"professional"`
	LightingAdjustment *string `json:"lightingAdjustment" example:"soft"`

	// Style Transfer fields
	StyleType               *StyleType       `json:"styleType" validate:"omitempty,oneof=ARTISTIC MOOD BRAND CREATIVE" example:"ARTISTIC"`
	StylePrompt             *string          `json:"stylePrompt" example:"vintage style"`
	StyleReferenceImageUrl  *string          `json:"styleReferenceImageUrl" example:"https://example.com/style.jpg"`
	StyleReferenceImageType *ImageSourceType `json:"styleReferenceImageType" validate:"omitempty,oneof=URL BASE64" example:"URL"`
	Intensity               *float64         `json:"intensity" validate:"omitempty,min=0,max=1" example:"0.7"`
	PreserveDetails         *bool            `json:"preserveDetails" example:"true"`
	BrandGuidelines         *string          `json:"brandGuidelines" example:"Brand specific guidelines"`

	// Scene Generation fields
	SceneType          *SceneType `json:"sceneType" validate:"omitempty,oneof=LIFESTYLE COMMERCIAL CONTEXT BRAND_ENVIRONMENT" example:"LIFESTYLE"`
	ScenePrompt        *string    `json:"scenePrompt" example:"Product in a cozy living room"`
	ContextDescription *string    `json:"contextDescription" example:"Modern home setting"`
	LifestyleSettings  *string    `json:"lifestyleSettings" example:"casual home environment"`
	BrandEnvironment   *string    `json:"brandEnvironment" example:"luxury brand setting"`
	LightingConditions *string    `json:"lightingConditions" example:"natural daylight"`
	TargetAudience     *string    `json:"targetAudience" example:"young professionals"`

	// Product Color Variants fields
	Prompt         *string `json:"prompt" example:"transform the product[1] color in black with white strip"`
	ItemsToRecolor *string `json:"itemsToRecolor" example:"shoe"`
	Color          *string `json:"color" example:"black"`

	// Product Variants fields
	VariantType          *VariantType `json:"variantType" validate:"omitempty,oneof=SIZE CONFIGURATION ANGLE BATCH" example:"SIZE"`
	VariantPrompt        *string      `json:"variantPrompt" example:"Different size variations"`
	NumberOfVariants     *int         `json:"numberOfVariants" validate:"omitempty,min=1,max=10" example:"3"`
	SizeVariations       []string     `json:"sizeVariations" example:"[\"small\", \"medium\", \"large\"]"`
	ConfigurationOptions []string     `json:"configurationOptions" example:"[\"open\", \"closed\", \"partial\"]"`
	AngleVariations      []string     `json:"angleVariations" example:"[\"front\", \"side\", \"back\"]"`
	ColorVariations      []string     `json:"colorVariations" example:"[\"red\", \"blue\", \"green\"]"`
	MaterialVariations   []string     `json:"materialVariations" example:"[\"metal\", \"wood\", \"plastic\"]"`
}
