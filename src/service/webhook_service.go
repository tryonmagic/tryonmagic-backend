package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/mbbipul/prodigyai-backend/src/config"
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/mbbipul/prodigyai-backend/src/pkg/downloader"
	"github.com/mbbipul/prodigyai-backend/src/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	s3ada "github.com/mbbipul/prodigyai-backend/src/internal/adapter/s3"
	socket "github.com/mbbipul/prodigyai-backend/src/internal/adapter/socketio"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WebhookService interface {
	FashnAIVTonWebhook(c *fiber.Ctx, taskId uuid.UUID, responseData FashnAIVtonResponseData) error
}

type webhookService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewWebhookService(db *gorm.DB, validate *validator.Validate) WebhookService {
	return &webhookService{
		Log:      utils.Log,
		DB:       db,
		Validate: validate,
	}
}

func (t *webhookService) FashnAIVTonWebhook(c *fiber.Ctx, taskId uuid.UUID, responseData FashnAIVtonResponseData) error {

	jobID := responseData.ID

	resultMap := map[string]interface{}{
		"jobId":          jobID,
		"status":         responseData.Status,
		"output":         responseData.Output,
		"outputImageUrl": responseData.Output,
		"error":          responseData.Error,
	}

	resultBytes, err := json.Marshal(resultMap)
	if err != nil {
		return fmt.Errorf("failed to marshal result data: %w", err)
	}

	resultData := (*datatypes.JSON)(&resultBytes)

	taskObj := new(model.Task)
	taskObj.ID = taskId
	taskObj.ResultData = resultData
	taskObj.Status = MapVtonJobStatusToTaskStatus(*responseData.Status)
	now := time.Now()
	taskObj.CompletedAt = &now

	if err := t.DB.WithContext(context.Background()).Model(taskObj).Clauses(clause.Returning{}).Updates(taskObj).Error; err != nil {
		fmt.Printf("Failed to update task: %+v", err)
		return err
	}
	err = socket.EmitEventToUser(taskObj.UserID.String(), "task-progress", map[string]any{
		"id":         taskObj.ID,
		"type":       taskObj.Type,
		"status":     taskObj.Status,
		"resultData": taskObj.ResultData,
	})
	if err != nil {
		t.Log.Errorf("Failed to emit FashnAIVTonWebhook task-progress event: %v", err)
	} else {
		t.Log.Infof("Emitted FashnAIVTonWebhook task-progress event to user: %s", taskObj.UserID.String())
	}

	go t.processOutputImagesAsync(taskObj, responseData.Output)

	return nil
}

func (t *webhookService) processOutputImagesAsync(task *model.Task, output *[]string) {
	ctx := context.Background()
	outputImageUrls := []string{}
	for _, outputUrl := range *output {
		outputImageUrl := strings.Split(outputUrl, "/")
		outputImageName := outputImageUrl[len(outputImageUrl)-1]

		bucket := config.S3BucketName
		key := "vton/" + task.ID.String() + "/" + outputImageName
		if err := SaveImageToS3(ctx, outputUrl, bucket, key); err != nil {
			t.Log.Errorf("Failed to save image to S3: %v", err)
		} else {
			outputImageUrls = append(outputImageUrls, "s3://"+bucket+"/"+key)
		}
	}
	resultDataMap := map[string]interface{}{}
	if err := json.Unmarshal(*task.ResultData, &resultDataMap); err != nil {
		t.Log.Errorf("Failed to unmarshal existing ResultData: %v", err)
		return
	}

	resultDataMap["outputImageUrl"] = outputImageUrls

	updatedResultBytes, err := json.Marshal(resultDataMap)
	if err != nil {
		t.Log.Errorf("Failed to marshal updated ResultData: %v", err)
		return
	}

	task.ResultData = (*datatypes.JSON)(&updatedResultBytes)
	if err := t.DB.WithContext(ctx).Model(task).Clauses(clause.Returning{}).Updates(task).Error; err != nil {
		t.Log.Errorf("Failed to update task with outputImageUrls: %v", err)
	}

}

func SaveImageToS3(ctx context.Context, imageUrl, bucket, key string) error {
	downloader := downloader.BuildDownloader(imageUrl)
	if downloader == nil {
		return fmt.Errorf("failed to build downloader for url: %s", imageUrl)
	}
	writer := manager.NewWriteAtBuffer([]byte{})
	s3Client := s3ada.GetS3Client()

	err := downloader.Download(imageUrl, writer)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	// upload the image to S3
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(writer.Bytes()),
	})

	return err
}
