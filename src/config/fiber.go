package config

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
)

func FiberConfig() fiber.Config {
	return fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		ServerHeader:  "Fiber",
		AppName:       "Fiber API",
		ErrorHandler:  utils.ErrorHandler,
		JSONEncoder:   sonic.Marshal,
		JSONDecoder:   sonic.Unmarshal,
		// Increase timeouts for long-running operations like image uploads
		ReadTimeout:  10 * time.Minute, // Allow 10 minutes for reading requests (uploads)
		WriteTimeout: 10 * time.Minute, // Allow 10 minutes for writing responses
		IdleTimeout:  15 * time.Minute, // Keep connections alive for 15 minutes
		// Allow larger request bodies for image uploads
		BodyLimit: 100 * 1024 * 1024, // 100MB max body size
		// Disable connection reuse to avoid race conditions with fasthttp context
		DisableKeepalive: false, // Keep this false for performance, but we handle race conditions in HTTP clients
		// Reduce concurrent connections to minimize race conditions
		Concurrency: 1024 * 256, // Default is 256 * 1024, reduce to 256 * 1024 for safety
	}
}
