package downloader

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type LocalDownloader struct {
}

func (ld LocalDownloader) Download(url string, w io.WriterAt) error {
	filePath := strings.Replace(url, string(LOCAL), "", 1)
	fmt.Printf("File path: %v", filePath)
	imgData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	w.WriteAt(imgData, 0)
	return nil
}

func NewLocalDownloader() LocalDownloader {
	return LocalDownloader{}
}
