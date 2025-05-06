package downloader

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	s3ada "github.com/mbbipul/prodigyai-backend/src/internal/adapter/s3"
	"github.com/mbbipul/prodigyai-backend/src/utils"
)

type S3Downloader struct {
	S3Client *s3.Client
}

func (downloader S3Downloader) Download(url string, w io.WriterAt) error {
	mDownloader := manager.NewDownloader(downloader.S3Client, func(d *manager.Downloader) {
		d.PartSize = 1 * 1024 * 1024 // 1MB per part
		d.PartBodyMaxRetries = 3
		d.Concurrency = 10
	})
	t := time.Now()
	bucket, key := utils.ParseS3URL(url)
	_, err := mDownloader.Download(context.TODO(), w, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Elapsed download time of file %s %v", key, time.Since(t))
	return nil
}

func NewS3Downloader() S3Downloader {
	s3Client := s3ada.GetS3Client()
	return S3Downloader{
		S3Client: s3Client,
	}
}
