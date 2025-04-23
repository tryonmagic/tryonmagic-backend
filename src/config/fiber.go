package config

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/prodigyai-backend/src/utils"
)

func FiberConfig() fiber.Config {
	return fiber.Config{
		Prefork:       IsProd,
		CaseSensitive: true,
		ServerHeader:  "Fiber",
		AppName:       "Fiber API",
		ErrorHandler:  utils.ErrorHandler,
		JSONEncoder:   sonic.Marshal,
		JSONDecoder:   sonic.Unmarshal,
	}
}
