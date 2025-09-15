package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpDownloader struct {
	Client     *http.Client
	SaveToFile bool
	Threads    int
}

func (dl HttpDownloader) Download(url string, w io.WriterAt) error {
	Info("Starting single threaded download...")

	// Create a context with timeout to avoid hanging requests
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := dl.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	w.WriteAt(body, 0)

	return nil
}

func Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func NewHttpDownloader() HttpDownloader {
	return HttpDownloader{
		Client: &http.Client{
			Timeout: 2 * time.Minute, // Set 2-minute timeout for image downloads
			Transport: &http.Transport{
				DisableKeepAlives:     true, // Disable keep-alives to avoid context conflicts
				DisableCompression:    false,
				MaxIdleConns:          0,
				MaxIdleConnsPerHost:   0,
				MaxConnsPerHost:       0,
				IdleConnTimeout:       0,
				ResponseHeaderTimeout: 30 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		SaveToFile: true,
		Threads:    4,
	}
}
