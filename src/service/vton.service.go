package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/config"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/pkg/downloader"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
	"github.com/sirupsen/logrus"
)

type GarmentCategory string

const (
	Tops      GarmentCategory = "tops"
	Bottoms   GarmentCategory = "bottoms"
	OnePieces GarmentCategory = "one-pieces"
)

type JobStatus string

const (
	StatusStarting   JobStatus = "starting"
	StatusInQueue    JobStatus = "in_queue"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

func MapVtonJobStatusToTaskStatus(status JobStatus) model.TaskStatus {
	switch status {
	case StatusStarting:
		return model.TaskStatusPending
	case StatusInQueue:
		return model.TaskStatusQueued
	case StatusProcessing:
		return model.TaskStatusProcessing
	case StatusCompleted:
		return model.TaskStatusSucceeded
	case StatusFailed:
		return model.TaskStatusFailed
	default:
		return model.TaskStatusPending
	}
}

type FashnAIVtonResponseData struct {
	ID     string     `json:"id"`
	Status *JobStatus `json:"status"`
	Output *[]string  `json:"output,omitempty"`
	Error  *string    `json:"error,omitempty"`
}

func MapOutfitTypeToGarmentCategory(outfitType model.OutfitType) (GarmentCategory, error) {
	switch outfitType {
	case model.OutfitTypeTop:
		return Tops, nil
	case model.OutfitTypeBottom:
		return Bottoms, nil
	case model.OutfitTypeFullBody:
		return OnePieces, nil
	default:
		return "", errors.New("invalid outfit type")
	}
}

type VirtualTryOnBase interface {
	VirtualTryOn(ctx context.Context, modelImageURL, garmentImageURL string, category GarmentCategory) (string, error)
}

type FashnAIVirtualTryOn struct {
	FashnBaseURL string
	APIKey       string
	Skip         bool
	Log          *logrus.Logger
}

func NewFashnAIVirtualTryOn(apiKey string, log *logrus.Logger) *FashnAIVirtualTryOn {
	return &FashnAIVirtualTryOn{
		FashnBaseURL: "https://api.fashn.ai/v1",
		APIKey:       apiKey,
		Skip:         false,
		Log:          log,
	}
}

func (f *FashnAIVirtualTryOn) VirtualTryOn(ctx context.Context, taskId uuid.UUID, vtonTask validation.VTONTaskInput) (*FashnAIVtonResponseData, error) {
	if f.Skip {
		return nil, fmt.Errorf("Virtual try-on skip not implemented")
	}

	category, err := MapOutfitTypeToGarmentCategory(vtonTask.OutfitType)
	if err != nil {
		return nil, err
	}

	garmentImageUrl, err := f.GetImageFromSource(ctx, vtonTask.GarmentImage)
	fmt.Println("presigned Garment image URL:", garmentImageUrl)
	if err != nil {
		f.Log.Errorf("Failed to get image URL for garment image: %v", err)
		return nil, fmt.Errorf("Garment image url is invalid %w", err)
	}
	modelImageUrl, err := f.GetImageFromSource(ctx, vtonTask.ModelImage)
	fmt.Println("presigned Model image URL:", modelImageUrl)
	if err != nil {
		f.Log.Errorf("Failed to get image URL for model image: %v", err)
		return nil, fmt.Errorf("Model image url is invalid %w", err)
	}

	payload := map[string]interface{}{
		"model_image":   modelImageUrl,
		"garment_image": garmentImageUrl,
		"category":      category,
		"num_samples":   vtonTask.NumSamples,
		"seed":          vtonTask.Seed,
		"mode":          vtonTask.ModelMode,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	jobUrl := f.FashnBaseURL + "/run"
	webhook_url := config.WebhookBaseURL + "/fashnai-vton/" + config.WebhookAPIKey + "/" + taskId.String()
	jobUrl += "?webhook_url=" + webhook_url
	req, err := http.NewRequest("POST", jobUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+f.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	f.Log.Infof("Response from fashnai: %v, string: %s", body, string(body))
	responseData := new(FashnAIVtonResponseData)
	if err := json.Unmarshal(body, responseData); err != nil {
		f.Log.Errorf("Failed to unmarshal response from fashnai: %v", err)
		return nil, err
	}

	return responseData, nil
}

func (f *FashnAIVirtualTryOn) GetImageFromSource(ctx context.Context, imageSource string) (string, error) {
	return downloader.GetImageFromSource(ctx, imageSource)
}

func (f *FashnAIVirtualTryOn) GetJobStatus(jobID string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", f.FashnBaseURL+"/status/"+jobID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+f.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, err
	}

	return responseData, nil
}

func (f *FashnAIVirtualTryOn) GetTryOnImage(jobID string) (string, error) {
	maxRetries := 3
	retryCount := 0

	for retryCount < maxRetries {
		responseData, err := f.GetJobStatus(jobID)
		if err != nil {
			return "", err
		}

		status, ok := responseData["status"].(string)
		if !ok {
			return "", errors.New("invalid response format")
		}

		if status == "failed" {
			return "", errors.New("try-on failed")
		} else if status == "completed" {
			output, ok := responseData["output"].([]interface{})
			if !ok || len(output) == 0 {
				return "", errors.New("invalid output format")
			}
			return output[0].(string), nil
		}

		time.Sleep(5 * time.Second)
		retryCount++
	}

	return "", errors.New("try-on processing took too long")
}
