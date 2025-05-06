package downloader

import (
	"fmt"
	"io"
	"log"
	"strings"
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
