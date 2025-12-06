package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/config"
	"github.com/mbbipul/tryonmagic-backend/src/pkg/downloader"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
	"github.com/sirupsen/logrus"
	"google.golang.org/genai"
)

// GenAIVirtualTryOn represents the GenAI-based virtual try-on service
type GenAIVirtualTryOn struct {
	Log *logrus.Logger
}

// NewGenAIVirtualTryOn creates a new GenAI virtual try-on service instance
func NewGenAIVirtualTryOn(log *logrus.Logger) *GenAIVirtualTryOn {
	return &GenAIVirtualTryOn{
		Log: log,
	}
}

// GenAIResponse represents the response structure for GenAI operations
type GenAIResponse struct {
	ImageData []byte `json:"imageData"`
	MimeType  string `json:"mimeType"`
}

// FileToPart converts a file (from URL or base64) to a GenAI Part
func (g *GenAIVirtualTryOn) FileToPart(ctx context.Context, fileSource string) (*genai.Part, error) {
	// Check if it's a base64 data URL
	if strings.HasPrefix(fileSource, "data:") {
		return g.base64ToPart(fileSource)
	}

	// Otherwise, treat it as a URL and download
	return g.urlToPart(ctx, fileSource)
}

// base64ToPart converts a base64 data URL to a GenAI Part
func (g *GenAIVirtualTryOn) base64ToPart(dataURL string) (*genai.Part, error) {
	// Parse data URL format: data:mime/type;base64,data
	parts := strings.Split(dataURL, ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid data URL format")
	}

	// Extract MIME type
	mimeInfo := parts[0]
	mimeType := "image/png" // default
	if strings.Contains(mimeInfo, ":") && strings.Contains(mimeInfo, ";") {
		start := strings.Index(mimeInfo, ":") + 1
		end := strings.Index(mimeInfo, ";")
		if start < end {
			mimeType = mimeInfo[start:end]
		}
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	return &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     data,
		},
	}, nil
}

// urlToPart downloads a file from URL and converts it to a GenAI Part
func (g *GenAIVirtualTryOn) urlToPart(ctx context.Context, url string) (*genai.Part, error) {
	// Use the existing downloader
	downloader := downloader.BuildDownloader(url)
	if downloader == nil {
		return nil, fmt.Errorf("failed to build downloader for URL: %s", url)
	}

	writer := manager.NewWriteAtBuffer([]byte{})
	err := downloader.Download(url, writer)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	// Determine MIME type from URL extension or HTTP headers
	mimeType := g.determineMimeType(url)

	return &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     writer.Bytes(),
		},
	}, nil
}

// determineMimeType determines MIME type from URL extension
func (g *GenAIVirtualTryOn) determineMimeType(url string) string {
	ext := strings.ToLower(filepath.Ext(url))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// Default to PNG if we can't determine
		mimeType = "image/png"
	}
	return mimeType
}

// HandleAPIResponse processes the GenAI API response and extracts image data
func (g *GenAIVirtualTryOn) HandleAPIResponse(response *genai.GenerateContentResponse, context string) (*GenAIResponse, error) {
	// Check for prompt blocking first
	if response.PromptFeedback != nil && response.PromptFeedback.BlockReason != "" {
		blockReason := response.PromptFeedback.BlockReason
		blockMessage := response.PromptFeedback.BlockReasonMessage
		errorMessage := fmt.Sprintf("Request was blocked. Reason: %s. %s", blockReason, blockMessage)
		g.Log.Error(errorMessage)
		return nil, fmt.Errorf("%s", errorMessage)
	}

	// Try to find the image part in candidates
	if len(response.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response for %s", context)
	}

	candidate := response.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts in response for %s", context)
	}

	// Look for image data in parts
	for _, part := range candidate.Content.Parts {
		if part.InlineData != nil && part.InlineData.Data != nil {
			mimeType := part.InlineData.MIMEType
			data := part.InlineData.Data

			g.Log.Infof("Received image data (%s) for %s", mimeType, context)
			return &GenAIResponse{
				ImageData: data,
				MimeType:  mimeType,
			}, nil
		}
	}

	// Check finish reason if no image found
	if candidate.FinishReason != "" && candidate.FinishReason != "STOP" {
		errorMessage := fmt.Sprintf("Image generation for %s stopped unexpectedly. Reason: %s. This often relates to safety settings.", context, candidate.FinishReason)
		g.Log.Error(errorMessage)
		return nil, fmt.Errorf("%s", errorMessage)
	}

	// Check for text feedback
	textFeedback := ""
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			textFeedback = part.Text
			break
		}
	}

	errorMessage := fmt.Sprintf("The AI model did not return an image for the %s.", context)
	if textFeedback != "" {
		errorMessage += fmt.Sprintf(" The model responded with text: \"%s\"", textFeedback)
	} else {
		errorMessage += " This can happen due to safety filters or if the request is too complex. Please try rephrasing your prompt to be more direct."
	}

	g.Log.Errorf("Model response did not contain an image part for %s", context)
	return nil, fmt.Errorf("%s", errorMessage)
}

// VirtualTryOnWithGemini performs virtual try-on using Google's Gemini model
func (g *GenAIVirtualTryOn) VirtualTryOnWithGemini(ctx context.Context, taskId uuid.UUID, vtonTask validation.VTONTaskInput) (*GenAIResponse, error) {
	// Create GenAI client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GEMINI_API_KEY,
	})
	if err != nil {
		g.Log.Errorf("Failed to create GenAI client: %v", err)
		return nil, fmt.Errorf("something went wrong")
	}

	// Convert model image to GenAI Part
	modelImagePart, err := g.FileToPart(ctx, vtonTask.ModelImage)
	if err != nil {
		g.Log.Errorf("Failed to convert model image to part: %v", err)
		return nil, fmt.Errorf("failed to process model image: %w", err)
	}

	// Convert garment image to GenAI Part
	garmentImagePart, err := g.FileToPart(ctx, vtonTask.GarmentImage)
	if err != nil {
		g.Log.Errorf("Failed to convert garment image to part: %v", err)
		return nil, fmt.Errorf("failed to process garment image: %w", err)
	}

	// Create the prompt
	prompt := `You are a virtual try-on AI expert. You will be given two images.
- Image 1 is the 'model' (the person).
- Image 2 is the 'garment' (the clothing item, which may be on another person or by itself).

Your task is to take the garment from Image 2 and realistically place it on the person in Image 1.

Editing Guidelines:
- Preserve the model's pose, face, body shape, and skin tone from Image 1.
- The model cloth should be removed completely.
- Realistically adapt the garment from Image 2 to fit the model's body and pose.
- The final image should be a photorealistic composition of the model wearing the new garment, with correct lighting, shadows, and drapes.
- The background should be the same as in Image 1.

Output: Return ONLY the final edited image. Do not return text.`

	textPart := genai.NewPartFromText(prompt)

	// Create content with all parts
	parts := []*genai.Part{modelImagePart, garmentImagePart, textPart}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	g.Log.Info("Sending model and garment images to the GenAI model...")

	// Generate content using Gemini 3 Pro Image Preview
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-pro-image-preview",
		contents,
		nil,
	)
	if err != nil {
		g.Log.Errorf("Failed to generate content: %v", err)
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	g.Log.Infof("Received response from GenAI model for try-on %v", result)

	// Handle the API response
	response, err := g.HandleAPIResponse(result, "virtual try-on")
	if err != nil {
		return nil, err
	}

	return response, nil
}
