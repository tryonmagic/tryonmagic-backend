package cloudinaryapi

import (
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func NewCloudinaryClient() *cloudinary.Cloudinary {
	cld, _ := cloudinary.New()
	cld.Config.URL.Secure = true
	// Increase API timeouts to avoid the 60s default cancellation inside the SDK
	// Values are in seconds
	cld.Config.API.Timeout = 480       // 8 minutes for form-based uploads (including remote URL uploads)
	cld.Config.API.UploadTimeout = 480 // 8 minutes for multipart uploads

	// Create a custom HTTP client with longer timeouts
	customClient := &http.Client{
		Timeout: 8 * time.Minute, // 8 minutes total timeout
		Transport: &http.Transport{
			ResponseHeaderTimeout: 8 * time.Minute,
			ExpectContinueTimeout: 30 * time.Second,
		},
	}

	// Create new uploader API with custom HTTP client
	uploaderAPI, _ := uploader.NewWithConfiguration(&cld.Config)
	uploaderAPI.Client = *customClient
	cld.Upload = *uploaderAPI

	return cld
}
