package cloudinaryapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v2"
	s3ada "github.com/mbbipul/tryonmagic-backend/src/internal/adapter/s3"
	"github.com/mbbipul/tryonmagic-backend/src/pkg/downloader"
)

type CloudinaryResponse struct {
	APIKey              string                `json:"api_key"`
	AssetFolder         string                `json:"asset_folder"`
	AssetID             string                `json:"asset_id"`
	Bytes               float64               `json:"bytes"`
	CreatedAt           string                `json:"created_at"`
	DisplayName         string                `json:"display_name"`
	Eager               []EagerTransformation `json:"eager"`
	ETag                string                `json:"etag"`
	Format              string                `json:"format"`
	Height              int                   `json:"height"`
	NotificationContext NotificationContext   `json:"notification_context"`
	NotificationType    string                `json:"notification_type"`
	OriginalFilename    string                `json:"original_filename"`
	Placeholder         bool                  `json:"placeholder"`
	PublicID            string                `json:"public_id"`
	RequestID           string                `json:"request_id"`
	ResourceType        string                `json:"resource_type"`
	SecureURL           string                `json:"secure_url"`
	SignatureKey        string                `json:"signature_key"`
	Tags                []string              `json:"tags"`
	Timestamp           string                `json:"timestamp"`
	Type                string                `json:"type"`
	URL                 string                `json:"url"`
	UseMTLS             bool                  `json:"use_mtls"`
	Version             float64               `json:"version"`
	VersionID           string                `json:"version_id"`
	Width               int                   `json:"width"`
}

type EagerTransformation struct {
	Reason         string `json:"reason"`
	Status         string `json:"status"`
	Transformation string `json:"transformation"`
}

type NotificationContext struct {
	TriggeredAt string                 `json:"triggered_at"`
	TriggeredBy map[string]interface{} `json:"triggered_by"`
}

func UploadImageToCloudinary(fileLocation string, destKey string, transformation string, webhookUrl string, parentCtx context.Context) (*string, error) {
	// Create a new context with a longer timeout specifically for Cloudinary uploads
	// Use 8 minutes timeout for uploads to handle large images (slightly less than server timeout)
	uploadCtx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	fmt.Printf("Starting Cloudinary upload for key: %s, transformation: %s\n", destKey, transformation)

	cld := NewCloudinaryClient()

	// Debug: Log the configured timeouts
	fmt.Printf("Cloudinary SDK Timeout: %d seconds, UploadTimeout: %d seconds\n",
		cld.Config.API.Timeout, cld.Config.API.UploadTimeout)
	fmt.Printf("Cloudinary HTTP Client Timeout: %v\n", cld.Upload.Client.Timeout)

	uploadParams := uploader.UploadParams{
		PublicID:             destKey,
		UniqueFilename:       api.Bool(false),
		Overwrite:            api.Bool(true),
		Eager:                transformation,
		EagerAsync:           api.Bool(true),
		EagerNotificationURL: webhookUrl,
	}

	// Debug: Log the webhook URL being sent to Cloudinary
	fmt.Printf("🔔 WEBHOOK DEBUG: Sending webhook URL to Cloudinary: %s\n", webhookUrl)

	startTime := time.Now()
	fmt.Printf("Cloudinary upload attempt started at: %v for key: %s\n", startTime, destKey)

	resp, err := cld.Upload.Upload(uploadCtx, fileLocation, uploadParams)
	uploadDuration := time.Since(startTime)

	if err != nil {
		fmt.Printf("Cloudinary upload failed after %v for key: %s, error: %v\n", uploadDuration, destKey, err)
		if strings.Contains(err.Error(), "context deadline exceeded") {
			fmt.Printf("Cloudinary upload timed out after %v for key: %s\n", uploadDuration, destKey)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Something went wrong. Please try again later!")
		}
		fmt.Printf("Error uploading image to Cloudinary (took %v): %v\n", uploadDuration, err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Something went wrong. Please try again later!")
	}

	fmt.Printf("Image uploaded successfully in %v: %s\n", uploadDuration, resp.SecureURL)
	return &resp.SecureURL, nil
}

func TransformImage(key string, transformation string) (string, error) {
	cld := NewCloudinaryClient()
	image, err := cld.Image(key)
	if err != nil {
		fmt.Println("error")
	}

	image.Transformation = transformation

	return image.String()

}

func SaveImageToS3(ctx context.Context, imageUrl, bucket, key string) error {
	downloader := downloader.BuildDownloader(imageUrl)
	if downloader == nil {
		return fmt.Errorf("failed to build downloader for url: %s", imageUrl)
	}
	writer := manager.NewWriteAtBuffer([]byte{})
	// s3Client := s3ada.GetS3Client()

	err := downloader.Download(imageUrl, writer)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	// upload the image to S3
	// _, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
	// 	Bucket: &bucket,
	// 	Key:    &key,
	// 	Body:   bytes.NewReader(writer.Bytes()),
	// })
	_, err = s3ada.UploadFileToS3(ctx, writer, bucket, key)

	return err
}
