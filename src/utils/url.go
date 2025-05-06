package utils

import "strings"

func ParseS3URL(s3URL string) (bucketName, key string) {
	cleanedImagePath := strings.Replace(s3URL, "s3://", "", 1)
	bucketName = strings.Split(cleanedImagePath, "/")[0]
	key = strings.Join(strings.Split(cleanedImagePath, "/")[1:], "/")
	return
}
