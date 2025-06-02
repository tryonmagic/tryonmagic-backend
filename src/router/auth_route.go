package router

import (
	m "github.com/mbbipul/tryonmagic-backend/src/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/config"
	"github.com/mbbipul/tryonmagic-backend/src/controller"
	"github.com/mbbipul/tryonmagic-backend/src/service"
)

func AuthRoutes(
	v1 fiber.Router, a service.AuthService, u service.UserService,
	t service.TokenService, e service.EmailService,
) {
	authController := controller.NewAuthController(a, u, t, e)
	config.GoogleConfig()

	auth := v1.Group("/auth")

	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/logout", authController.Logout)
	auth.Post("/refresh-tokens", authController.RefreshTokens)
	auth.Post("/forgot-password", authController.ForgotPassword)
	auth.Post("/reset-password", authController.ResetPassword)
	auth.Post("/send-verification-email", m.Auth(a, u), authController.SendVerificationEmail)
	auth.Post("/verify-email", authController.VerifyEmail)
	auth.Get("/google", authController.GoogleLogin)
	auth.Get("/google-callback", authController.GoogleCallback)
	auth.Post("/shopify-auth", authController.ShopifyAuth)
}
