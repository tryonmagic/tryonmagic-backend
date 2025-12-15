package config

import (
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/spf13/viper"
)

var (
	IsProd                          bool
	AppHost                         string
	AppPort                         int
	DBHost                          string
	DBUser                          string
	DBPassword                      string
	DBName                          string
	DBPort                          int
	JWTSecret                       string
	JWTAccessExp                    int
	JWTInternalExp                  int
	JWTRefreshExp                   int
	JWTResetPasswordExp             int
	JWTVerifyEmailExp               int
	SMTPHost                        string
	SMTPPort                        int
	SMTPUsername                    string
	SMTPPassword                    string
	EmailFrom                       string
	GoogleClientID                  string
	GoogleClientSecret              string
	RedirectURL                     string
	ShopifyAPIKey                   string
	ShopifyAPISecret                string
	ShopifyAPPScopes                string
	RabbitMQConnectionUrl           string
	AITaskExchangeName              string = "ai.task.exchange"
	RemoveBackgroundTaskQueueName   string = "queue.background_removal"
	RemoveBackgroundTaskRoutingKey  string = "task.background_removal"
	ReplaceBackgroundTaskQueueName  string = "queue.background_replacement"
	ReplaceBackgroundTaskRoutingKey string = "task.background_replacement"
	GenerateAIModelTaskQueueName    string = "queue.generate_ai_model"
	GenerateAIModelTaskRoutingKey   string = "task.generate_ai_model"
	ExpandImageTaskQueueName        string = "queue.expand_image"
	ExpandImageTaskRoutingKey       string = "task.expand_image"
	VTONTaskQueueName               string = "queue.vton"
	VTONTaskRoutingKey              string = "task.vton"
	UpscaleTaskQueueName            string = "queue.upscale"
	UpscaleTaskRoutingKey           string = "task.upscale"
	EditImageTaskQueueName          string = "queue.edit_image"
	EditImageTaskRoutingKey         string = "task.edit_image"
	ProductImageEditTaskQueueName   string = "queue.product_image_editing"
	ProductImageEditTaskRoutingKey  string = "task.product_image_editing"
	GrpcServerPort                  int32
	WebhookAPIKey                   string
	S3AccessKey                     string
	S3AccessID                      string
	S3Region                        string
	S3BucketName                    string
	S3Endpoint                      string
	WebhookBaseURL                  string
	FASHN_AI_API_KEY                string
	GEMINI_API_KEY                  string
	VTON_SERVICE                    string
	CloudinaryURL                   string
)

func init() {
	loadConfig()

	// server configuration
	IsProd = viper.GetString("APP_ENV") == "prod"
	AppHost = viper.GetString("APP_HOST")
	AppPort = viper.GetInt("APP_PORT")

	// database configuration
	DBHost = viper.GetString("DB_HOST")
	DBUser = viper.GetString("DB_USER")
	DBPassword = viper.GetString("DB_PASSWORD")
	DBName = viper.GetString("DB_NAME")
	DBPort = viper.GetInt("DB_PORT")

	// jwt configuration
	JWTSecret = viper.GetString("JWT_SECRET")
	JWTAccessExp = viper.GetInt("JWT_ACCESS_EXP_MINUTES")
	JWTInternalExp = viper.GetInt("JWT_INTERNAL_ACCESS_EXP_DAYS")
	JWTRefreshExp = viper.GetInt("JWT_REFRESH_EXP_DAYS")
	JWTResetPasswordExp = viper.GetInt("JWT_RESET_PASSWORD_EXP_MINUTES")
	JWTVerifyEmailExp = viper.GetInt("JWT_VERIFY_EMAIL_EXP_MINUTES")

	// SMTP configuration
	SMTPHost = viper.GetString("SMTP_HOST")
	SMTPPort = viper.GetInt("SMTP_PORT")
	SMTPUsername = viper.GetString("SMTP_USERNAME")
	SMTPPassword = viper.GetString("SMTP_PASSWORD")
	EmailFrom = viper.GetString("EMAIL_FROM")

	// oauth2 configuration
	GoogleClientID = viper.GetString("GOOGLE_CLIENT_ID")
	GoogleClientSecret = viper.GetString("GOOGLE_CLIENT_SECRET")
	RedirectURL = viper.GetString("REDIRECT_URL")
	ShopifyAPIKey = viper.GetString("SHOPIFY_API_KEY")
	ShopifyAPISecret = viper.GetString("SHOPIFY_API_SECRET")
	ShopifyAPPScopes = viper.GetString("SHOPIFY_APP_SCOPES")
	RabbitMQConnectionUrl = viper.GetString("RABBITMQ_CONNECTION_URI")
	GrpcServerPort = int32(viper.GetInt("GRPC_SERVER_PORT"))
	WebhookAPIKey = viper.GetString("WEBHOOK_API_KEY")
	S3AccessKey = viper.GetString("S3_ACCESS_KEY")
	S3AccessID = viper.GetString("S3_ACCESS_ID")
	S3Region = viper.GetString("S3_REGION")
	S3BucketName = viper.GetString("S3_BUCKET")
	S3Endpoint = viper.GetString("S3_ENDPOINT")
	WebhookBaseURL = viper.GetString("WEBHOOK_BASE_URL")
	FASHN_AI_API_KEY = viper.GetString("FASHN_AI_API_KEY")
	GEMINI_API_KEY = viper.GetString("GEMINI_API_KEY")
	VTON_SERVICE = viper.GetString("VTON_SERVICE")
	if VTON_SERVICE == "" {
		VTON_SERVICE = "gemini"
	}
	CloudinaryURL = viper.GetString("CLOUDINARY_URL")
}

func loadConfig() {
	// First priority: Read from process environment variables
	viper.AutomaticEnv()

	// Second priority: Try to read from .env file as fallback
	configPaths := []string{
		"./",     // For app
		"../../", // For test folder
	}

	for _, path := range configPaths {
		viper.SetConfigFile(path + ".env")

		if err := viper.ReadInConfig(); err == nil {
			utils.Log.Infof("Config file loaded from %s (as fallback)", path)
			break
		}
	}

	utils.Log.Info("Environment variables prioritized from process environment")
}
