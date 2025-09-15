package s3

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg "github.com/mbbipul/tryonmagic-backend/src/config"
)

var s3Lock = &sync.Mutex{}

var s3ClientInstance *s3.Client

func GetS3Client() *s3.Client {
	ACCESS_KEY := cfg.S3AccessID
	SECRET_KEY := cfg.S3AccessKey

	if s3ClientInstance == nil {
		s3Lock.Lock()
		defer s3Lock.Unlock()
		if s3ClientInstance == nil {
			fmt.Println("Creating S3Client single instance now.")
			sdkConfig, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("ap-southeast-1"), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(ACCESS_KEY, SECRET_KEY, "")),
				config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
					func(service, region string, options ...interface{}) (aws.Endpoint, error) {
						return aws.Endpoint{
							URL: "https://s3.ap-southeast-1.wasabisys.com",
						}, nil
					},
				)))
			if err != nil {
				fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
				fmt.Println(err)
				panic(err)
			}
			s3ClientInstance = s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
				o.UsePathStyle = true
			})
		}
	}

	return s3ClientInstance
}

func UploadFileToS3(ctx context.Context, w *manager.WriteAtBuffer, bucket, key string) (*manager.UploadOutput, error) {

	// Create an S3 client using the existing GetS3Client function
	client := GetS3Client()

	// Create an uploader using the S3 client
	uploader := manager.NewUploader(client)

	return uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(w.Bytes()),
	})

}

func UploadBytesToS3(ctx context.Context, data []byte, bucket, key string) (*manager.UploadOutput, error) {

	// Create an S3 client using the existing GetS3Client function
	client := GetS3Client()

	// Create an uploader using the S3 client
	uploader := manager.NewUploader(client)

	return uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

}
