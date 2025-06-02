package downloader

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mbbipul/tryonmagic-backend/src/internal/adapter/s3"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
)

type FILE_SOURCE_PREFIX string

const (
	LOCAL FILE_SOURCE_PREFIX = "local://"
	HTTP  FILE_SOURCE_PREFIX = "http"
	S3    FILE_SOURCE_PREFIX = "s3://"
)

type Downloader interface {
	Download(url string, w io.WriterAt) error
}

func NewDownloader(downloader Downloader) Downloader {
	return downloader
}

func BuildDownloader(url string) Downloader {
	fmt.Printf("BuildDownloader url: %s\n", url)
	if strings.HasPrefix(url, string(LOCAL)) {
		fmt.Printf("f")
		return NewLocalDownloader()
	}
	if strings.HasPrefix(url, string(HTTP)) {
		return NewHttpDownloader()
	}
	if strings.HasPrefix(url, string(S3)) {
		return NewS3Downloader()
	}
	log.Printf("url not match with any downloader prefix")
	return nil
}

func GetImageFromSource(ctx context.Context, imageSource string) (string, error) {

	if strings.HasPrefix(imageSource, "s3://") {
		return getImagePresignedURL(ctx, imageSource)
	} else if strings.HasPrefix(imageSource, "http://") || strings.HasPrefix(imageSource, "https://") {
		return imageSource, nil
	} else if strings.HasPrefix(imageSource, "data:image/") {
		return imageSource, nil
	}

	if _, err := os.Stat(imageSource); err == nil {
		file, err := os.Open(imageSource)
		if err != nil {
			log.Fatalf("failed to open file: %v", err)
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(fileBytes), nil
	}
	return "", fmt.Errorf("invalid image source: %s", imageSource)
}

func getImagePresignedURL(ctx context.Context, imagePath string) (string, error) {
	fmt.Println("Getting presigned URL for image:", imagePath)
	s3Presigner := s3.NewPresigner()
	bucketName, imageKey := utils.ParseS3URL(imagePath)
	fmt.Printf("getImagePresignedURL: Bucket: %s, Key: %s\n", bucketName, imageKey)

	url, err := s3Presigner.GetObject(ctx, bucketName, imageKey, 3600*24*4)
	fmt.Printf("getImagePresignedURL: Presigned URL: %s\n", url)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL for image: %w", err)
	}
	return url, nil
}
