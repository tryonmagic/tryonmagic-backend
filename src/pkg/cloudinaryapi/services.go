package cloudinaryapi

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	s3ada "github.com/mbbipul/prodigyai-backend/src/internal/adapter/s3"
	"github.com/mbbipul/prodigyai-backend/src/pkg/downloader"
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

func UploadImageToCloudinary(fileLocation string, destKey string, transformation string, webhookUrl string, ctx context.Context) (*string, error) {

	cld := Credentials()
	uploadParams := uploader.UploadParams{
		PublicID:        destKey,
		UniqueFilename:  api.Bool(false),
		Overwrite:       api.Bool(true),
		Async:           api.Bool(true),
		NotificationURL: webhookUrl,
		Eager:           transformation,
	}
	resp, err := cld.Upload.Upload(ctx, fileLocation, uploadParams)
	if err != nil {
		fmt.Println("Error uploading image:", err)
		return nil, err
	}
	fmt.Println("Image uploaded successfully:", resp.SecureURL)

	return &resp.SecureURL, nil
}

func TransformImage(key string, transformation string) (string, error) {
	cld := Credentials()
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
