package downloader

import (
	"fmt"
	"io"
	"net/http"
)

type HttpDownloader struct {
	Client     *http.Client
	SaveToFile bool
	Threads    int
}

func (dl HttpDownloader) Download(url string, w io.WriterAt) error {
	Info("Starting single threaded download...")
	req, err := http.NewRequest("GET", url, nil)
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
		Client:     &http.Client{},
		SaveToFile: true,
		Threads:    4,
	}
}
