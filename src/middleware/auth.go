package middleware

import (
	"strings"

	"github.com/mbbipul/prodigyai-backend/src/config"
	"github.com/mbbipul/prodigyai-backend/src/model"
	"github.com/mbbipul/prodigyai-backend/src/service"
	"github.com/mbbipul/prodigyai-backend/src/utils"
	"github.com/mbbipul/prodigyai-backend/src/validation"

	"github.com/gofiber/fiber/v2"
)

func verifyShopifyAuth(c *fiber.Ctx, authService service.AuthService, token string) (*model.User, error) {

	if token == "" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Please authenticate")
	}

	user, err := authService.ShopifyAuth(c, &validation.ShopifyAuth{
		SessionToken: token,
	}, true, false)
	if err != nil || user == nil {
		utils.Log.Errorf("Failed to verify shopify auth: %+v", err)
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Please authenticate")
	}

	return user, nil

}
func Auth(authService service.AuthService, userService service.UserService, requiredRights ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("x-shopify-session-token")
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		var user *model.User = nil
		if token != "" {
			_user, err := verifyShopifyAuth(c, authService, token)
			if err != nil {
				return err
			}
			user = _user
		} else {
			authHeader = c.Get("Authorization")
			token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

			if token == "" {
				return fiber.NewError(fiber.StatusUnauthorized, "Please authenticate")
			}

			userID, err := utils.VerifyToken(token, config.JWTSecret, config.TokenTypeAccess)
			if err != nil {
				return fiber.NewError(fiber.StatusUnauthorized, "Please authenticate")
			}

			_user, err := userService.GetUserByID(c, userID)
			if err != nil {
				return fiber.NewError(fiber.StatusUnauthorized, "Please authenticate")
			}
			user = _user
		}
		if user == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Please authenticate")
		}

		c.Locals("user", user)

		if len(requiredRights) > 0 {
			userRights, hasRights := config.RoleRights[user.Role]
			if (!hasRights || !hasAllRights(userRights, requiredRights)) && c.Params("userId") != user.ID.String() {
				return fiber.NewError(fiber.StatusForbidden, "You don't have permission to access this resource")
			}
		}

		return c.Next()
	}
}

func hasAllRights(userRights, requiredRights []string) bool {
	rightSet := make(map[string]struct{}, len(userRights))
	for _, right := range userRights {
		rightSet[right] = struct{}{}
	}

	for _, right := range requiredRights {
		if _, exists := rightSet[right]; !exists {
			return false
		}
	}
	return true
}
