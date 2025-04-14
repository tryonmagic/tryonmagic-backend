package validation

type SaveFile struct {
	Key      string `json:"key" validate:"required,max=255" example:"unique-key"`
	FullPath string `json:"fullPath" validate:"required,max=1024" example:"/path/to/file"`
	URL      string `json:"url" validate:"required,url,max=2048" example:"https://example.com/file"`
	Name     string `json:"name" validate:"required,max=255" example:"file_name"`
	Type     string `json:"type" validate:"required,max=50" example:"image/png"`
	Size     int    `json:"size" validate:"required,min=1" example:"1024"`
	Width    int    `json:"width" validate:"required,min=1" example:"1920"`
	Height   int    `json:"height" validate:"required,min=1" example:"1080"`
	Ext      string `json:"ext" validate:"required,max=10" example:".png"`
	Mime     string `json:"mime" validate:"required,max=50" example:"image/png"`
	// ThumbHash *[]byte `json:"thumbHash" validate:"omitempty" example:"[0,1,2,3]"`
}
