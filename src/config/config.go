package config

import (
	"github.com/mbbipul/prodigyai-backend/src/utils"
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
	GrpcServerPort                  int32
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
}

func loadConfig() {
	configPaths := []string{
		"./",     // For app
		"../../", // For test folder
	}

	for _, path := range configPaths {
		viper.SetConfigFile(path + ".env")

		if err := viper.ReadInConfig(); err == nil {
			utils.Log.Infof("Config file loaded from %s", path)
			return
		}
	}

	utils.Log.Error("Failed to load any config file")
}
