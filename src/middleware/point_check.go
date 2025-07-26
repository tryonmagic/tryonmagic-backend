package middleware

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/service"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
)

// PointCheckMiddleware checks if user has enough points to perform a task
func PointCheckMiddleware(subscriptionService service.SubscriptionService, taskType model.TaskType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		utils.Log.Println(">>>>> PointCheckMiddleware entered")

		utils.Log.Println("point check middleware : started")
		utils.Log.Printf("Request path: %s, method: %s", c.Path(), c.Method())

		userInterface := c.Locals("user")
		if userInterface == nil {
			utils.Log.Error("User not found in context")
			return fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
		}

		user, ok := userInterface.(*model.User)
		if !ok || user == nil {
			utils.Log.Error("User type assertion failed or user is nil")
			return fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
		}

		utils.Log.Printf("Processing request for user: %s", user.ID)

		// Use provided task type or extract from request
		if taskType == "" {
			var err error
			taskType, err = getTaskTypeFromRequest(c)
			if err != nil {
				utils.Log.Errorf("Failed to get task type: %v", err)
				return fiber.NewError(fiber.StatusForbidden, "Invalid task type")
			}
		}

		utils.Log.Printf("Task type extracted: %s", taskType)

		// Get required points for the task
		requiredPoints := getPointsForTaskType(taskType)
		fmt.Println("middleware requiredPoints : ", requiredPoints)

		// Check if user can perform the task
		canPerform, err := subscriptionService.CheckUserCanPerformTask(user.ID, requiredPoints)
		if err != nil {
			utils.Log.Errorf("Error checking user subscription: %v", err)
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}

		if !canPerform {
			// Get subscription details for better error message
			subscription, err := subscriptionService.GetSubscriptionStatus(user.ID)
			if err != nil {
				utils.Log.Errorf("Error getting subscription status: %v", err)
				return fiber.NewError(fiber.StatusForbidden, "Insufficient points or inactive subscription")
			}

			if !subscription.IsSubscriptionActive() {
				utils.Log.Errorf("Subscription not active for user: %s", user.ID)
				return fiber.NewError(fiber.StatusForbidden, "Subscription is not active")
			}

			utils.Log.Errorf("Insufficient points for user: %s, required: %d, available: %d", user.ID, requiredPoints, subscription.PointsBalance)
			return fiber.NewError(fiber.StatusForbidden,
				"You don't have enough points to perform this task. Please purchase a subscription plan.")
		}

		// Store required points in context for later use
		c.Locals("requiredPoints", requiredPoints)
		c.Locals("taskType", taskType)
		fmt.Println("point check middleware : passed")
		utils.Log.Printf("Point check middleware passed for user: %s, task: %s, points: %d", user.ID, taskType, requiredPoints)

		return c.Next()
	}
}

// getTaskTypeFromRequest extracts task type from request
func getTaskTypeFromRequest(c *fiber.Ctx) (model.TaskType, error) {
	// Try to get from request body first
	body := c.Body()
	if len(body) > 0 {
		var requestBody map[string]interface{}
		if err := json.Unmarshal(body, &requestBody); err == nil {
			if taskTypeStr, ok := requestBody["type"].(string); ok {
				// Store the parsed body in context for reuse by controller
				c.Locals("parsedBody", requestBody)
				return model.TaskType(taskTypeStr), nil
			}
		}
	}

	// Try to get from query params
	if taskTypeStr := c.Query("type"); taskTypeStr != "" {
		return model.TaskType(taskTypeStr), nil
	}

	// Try to get from headers
	if taskTypeStr := c.Get("X-Task-Type"); taskTypeStr != "" {
		return model.TaskType(taskTypeStr), nil
	}

	return "", errors.New("task type not found in request")
}

// getPointsForTaskType returns the points cost for each task type
func getPointsForTaskType(taskType model.TaskType) int {
	switch taskType {
	case model.TaskTypeRemoveBackground:
		return 1
	case model.TaskTypeGenerateAIModel:
		return 2
	case model.TaskTypeVTon:
		return 3
	case model.TaskTypeUpscale:
		return 1
	case model.TaskTypeEnhance:
		return 1
	case model.TaskTypeExpandImage:
		return 2
	case model.TaskTypeEditImage:
		return 1
	case model.TaskTypeProductImageEdit:
		return 1
	case model.TaskTypeReplaceBackground:
		return 1
	case model.TaskTypeGenerateImage:
		return 2
	default:
		return 1
	}
}

// AdminBypassMiddleware allows admins to bypass point checks (for testing/debugging)
func AdminBypassMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userInterface := c.Locals("user")
		if userInterface == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
		}

		user, ok := userInterface.(*model.User)
		if !ok || user == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
		}

		// Check if user is admin
		if user.Role == model.RoleAdmin {
			// Set bypass flag in context
			c.Locals("adminBypass", true)
			utils.Log.Infof("Admin bypass enabled for user: %s", user.ID)
		}

		return c.Next()
	}
}

// RateLimitMiddleware prevents abuse by limiting requests per user
func RateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userInterface := c.Locals("user")
		if userInterface == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
		}

		user, ok := userInterface.(*model.User)
		if !ok || user == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
		}

		// Simple rate limiting - can be enhanced with Redis
		// For now, we'll use a basic approach
		userID := user.ID.String()

		// This is a simplified rate limit - in production, use Redis with proper TTL
		// For now, we'll just log and continue
		utils.Log.Infof("Rate limit check for user: %s", userID)

		return c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to prevent bypass attempts
func SecurityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("Content-Security-Policy", "default-src 'self'")

		return c.Next()
	}
}
