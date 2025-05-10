package cloudinaryapi

import (
	"github.com/cloudinary/cloudinary-go/v2"
)

func Credentials() *cloudinary.Cloudinary {
	cld, _ := cloudinary.New()
	cld.Config.URL.Secure = true
	return cld
}
