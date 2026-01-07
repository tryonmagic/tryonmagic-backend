package service

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/mbbipul/tryonmagic-backend/src/config"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/mbbipul/tryonmagic-backend/src/pkg/shopify"
	"github.com/mbbipul/tryonmagic-backend/src/response"
	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthService interface {
	Register(c *fiber.Ctx, req *validation.Register) (*model.User, error)
	Login(c *fiber.Ctx, req *validation.Login) (*model.User, error)
	Logout(c *fiber.Ctx, req *validation.Logout) error
	RefreshAuth(c *fiber.Ctx, req *validation.RefreshToken) (*response.Tokens, error)
	ResetPassword(c *fiber.Ctx, query *validation.Token, req *validation.UpdatePassOrVerify) error
	VerifyEmail(c *fiber.Ctx, query *validation.Token) error
	ShopifyAuth(c *fiber.Ctx, payload *validation.ShopifyAuth, isOnline bool, store bool) (*model.User, error)
}

type authService struct {
	Log          *logrus.Logger
	DB           *gorm.DB
	Validate     *validator.Validate
	UserService  UserService
	TokenService TokenService
}

func NewAuthService(
	db *gorm.DB, validate *validator.Validate, userService UserService, tokenService TokenService,
) AuthService {
	return &authService{
		Log:          utils.Log,
		DB:           db,
		Validate:     validate,
		UserService:  userService,
		TokenService: tokenService,
	}
}

func (s *authService) Register(c *fiber.Ctx, req *validation.Register) (*model.User, error) {
	s.Log.Info(req.Role)
	if err := s.Validate.Struct(req); err != nil {
		s.Log.Error(err)
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.Log.Errorf("Failed hash password: %+v", err)
		return nil, err
	}

	user := &model.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      req.Role,
	}

	result := s.DB.WithContext(context.Background()).Create(user)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return nil, fiber.NewError(fiber.StatusConflict, "Email already taken")
	}

	if result.Error != nil {
		s.Log.Errorf("Failed create user: %+v", result.Error)
	}

	return user, result.Error
}

func (s *authService) Login(c *fiber.Ctx, req *validation.Login) (*model.User, error) {
	if err := s.Validate.Struct(req); err != nil {
		return nil, err
	}

	user, err := s.UserService.GetUserByEmail(c, req.Email)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid email or password")
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid email or password")
	}

	return user, nil
}

func (s *authService) Logout(c *fiber.Ctx, req *validation.Logout) error {
	if err := s.Validate.Struct(req); err != nil {
		return err
	}

	token, err := s.TokenService.GetTokenByUserID(c, req.RefreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Token not found")
	}

	err = s.TokenService.DeleteToken(c, config.TokenTypeRefresh, token.UserID.String())

	return err
}

func (s *authService) RefreshAuth(c *fiber.Ctx, req *validation.RefreshToken) (*response.Tokens, error) {
	if err := s.Validate.Struct(req); err != nil {
		return nil, err
	}

	token, err := s.TokenService.GetTokenByUserID(c, req.RefreshToken)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Please authenticate")
	}

	user, err := s.UserService.GetUserByID(c, token.UserID.String())
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Please authenticate")
	}

	newTokens, err := s.TokenService.GenerateAuthTokens(c, user)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return newTokens, err
}

func (s *authService) ResetPassword(c *fiber.Ctx, query *validation.Token, req *validation.UpdatePassOrVerify) error {
	if err := s.Validate.Struct(query); err != nil {
		return err
	}

	userID, err := utils.VerifyToken(query.Token, config.JWTSecret, config.TokenTypeResetPassword)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid Token")
	}

	user, err := s.UserService.GetUserByID(c, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Password reset failed")
	}

	if errUpdate := s.UserService.UpdatePassOrVerify(c, req, user.ID.String()); errUpdate != nil {
		return errUpdate
	}

	if errToken := s.TokenService.DeleteToken(c, config.TokenTypeResetPassword, user.ID.String()); errToken != nil {
		return errToken
	}

	return nil
}

func (s *authService) VerifyEmail(c *fiber.Ctx, query *validation.Token) error {
	if err := s.Validate.Struct(query); err != nil {
		return err
	}

	userID, err := utils.VerifyToken(query.Token, config.JWTSecret, config.TokenTypeVerifyEmail)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid Token")
	}

	user, err := s.UserService.GetUserByID(c, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Verify email failed")
	}

	if errToken := s.TokenService.DeleteToken(c, config.TokenTypeVerifyEmail, user.ID.String()); errToken != nil {
		return errToken
	}

	updateBody := &validation.UpdatePassOrVerify{
		EmailVerified: true,
	}

	if errUpdate := s.UserService.UpdatePassOrVerify(c, updateBody, user.ID.String()); errUpdate != nil {
		return errUpdate
	}

	return nil
}

func (s *authService) ShopifyAuth(c *fiber.Ctx, payload *validation.ShopifyAuth, isOnline bool, store bool) (*model.User, error) {
	s.Log.Infof("ShopifyAuth payload: %v", payload)
	s.Log.Infof("shopify secret: %v", config.ShopifyAPISecret)
	val, err := utils.DecodeJWTToken(payload.SessionToken, config.ShopifyAPISecret)
	if err != nil {
		s.Log.Errorf("Error decoding token: %v", err)
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid session token")
	}

	if val == nil {
		s.Log.Errorf("Decoded token claims are nil")
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid session token claims")
	}

	destInterface, exists := val["dest"]
	if !exists {
		s.Log.Errorf("Missing 'dest' claim in session token")
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid session token: missing destination")
	}

	dest, ok := destInterface.(string)
	if !ok {
		s.Log.Errorf("Invalid 'dest' claim type in session token: %T", destInterface)
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid session token: invalid destination format")
	}

	shop := strings.Replace(dest, "https://", "", 1)
	c.Locals("shop", shop)

	session, err := shopify.TokenExchange(shop, payload.SessionToken)
	if err != nil {
		s.Log.Errorf("Error exchanging token: %v", err)
		return nil, err
	}
	if !store {
		shopifySession := new(model.ShopifySession)

		// Use session ID to find the specific session (session ID includes shop + shopify_user_id)
		result := s.DB.WithContext(context.Background()).Where("id = ?", session.ID).Preload("User").First(shopifySession)

		if result.Error != nil {
			s.Log.Errorf("Failed to get shopify session: %+v", result.Error)
			return nil, result.Error
		}
		s.Log.Printf("Shopify Session user: %v", shopifySession.User)
		return shopifySession.User, nil
	}
	if session.OnlineAccessInfo != nil {
		sessionAssociatedUser := session.OnlineAccessInfo.AssociatedUser
		hashedPassword, err := utils.HashPassword(session.ID)
		if err != nil {
			s.Log.Errorf("Failed hash password: %+v", err)
			return nil, err
		}
		// Set shop and shopify_user_id to uniquely identify Shopify users
		shop := session.Shop
		shopifyUserID := sessionAssociatedUser.ID
		user := model.User{
			FirstName:     sessionAssociatedUser.FirstName,
			LastName:      sessionAssociatedUser.LastName,
			Email:         sessionAssociatedUser.Email,
			Role:          model.RoleAdmin,
			EmailVerified: sessionAssociatedUser.EmailVerified,
			Password:      hashedPassword,
			Shop:          &shop,
			ShopifyUserID: &shopifyUserID,
		}
		shopifySession := model.ShopifySession{
			ID:          session.ID,
			AccessToken: session.AccessToken,
			Expires:     &session.Expires,
			IsOnline:    session.IsOnline,
			Scope:       session.Scope,
			State:       session.State,
			APIKey:      config.ShopifyAPIKey,
			Shop:        session.Shop,
		}
		onlineAccessInfo := model.ShopifyOnlineAccessInfo{
			ShopifySessionID:    session.ID,
			ExpiresIn:           session.OnlineAccessInfo.ExpiresIn,
			AssociatedUserScope: session.OnlineAccessInfo.AssociatedUserScope,
		}
		associatedUser := model.ShopifyAssociatedUser{
			ShopifyUserID: sessionAssociatedUser.ID,
			FirstName:     sessionAssociatedUser.FirstName,
			LastName:      sessionAssociatedUser.LastName,
			Email:         sessionAssociatedUser.Email,
			AccountOwner:  sessionAssociatedUser.AccountOwner,
			Locale:        sessionAssociatedUser.Locale,
			Collaborator:  sessionAssociatedUser.Collaborator,
			EmailVerified: sessionAssociatedUser.EmailVerified,
		}

		err = s.DB.Transaction(func(tx *gorm.DB) error {
			// For Shopify users, use shop + shopify_user_id as unique identifier
			// First try to find existing user by shop + shopify_user_id
			var existingUser model.User
			findResult := tx.Where("shop = ? AND shopify_user_id = ?", shop, shopifyUserID).First(&existingUser)

			if findResult.Error == nil {
				// User exists, update it
				existingUser.FirstName = sessionAssociatedUser.FirstName
				existingUser.LastName = sessionAssociatedUser.LastName
				existingUser.Email = sessionAssociatedUser.Email
				existingUser.EmailVerified = sessionAssociatedUser.EmailVerified
				existingUser.Password = hashedPassword // Update password in case session ID changed
				if err := tx.Save(&existingUser).Error; err != nil {
					s.Log.Errorf("Failed to update user: %+v", err)
					return err
				}
				user = existingUser
			} else if errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
				// User doesn't exist, create new one
				if err := tx.Create(&user).Error; err != nil {
					s.Log.Errorf("Failed to store user: %+v", err)
					return err
				}
			} else {
				s.Log.Errorf("Failed to find user: %+v", findResult.Error)
				return findResult.Error
			}
			s.Log.Infof("User ID: %s", user.ID)
			shopifySession.UserID = user.ID
			if err := tx.Clauses(clause.Returning{}, clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"access_token", "expires", "is_online", "scope", "state", "api_key", "shop", "updated_at"}),
			}).Create(&shopifySession).Error; err != nil {
				s.Log.Errorf("Failed to store shopify session: %+v", err)
				return err
			}
			if err := tx.Clauses(clause.Returning{}, clause.OnConflict{
				Columns:   []clause.Column{{Name: "shopify_session_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"expires_in", "associated_user_scope", "updated_at"}),
			}).Create(&onlineAccessInfo).Error; err != nil {
				s.Log.Errorf("Failed to store online access info: %+v", err)
				return err
			}

			associatedUser.OnlineAccessInfoID = &onlineAccessInfo.ID
			s.Log.Infof("Associated User: %v", associatedUser)
			if err := tx.Clauses(clause.Returning{}, clause.OnConflict{
				Columns:   []clause.Column{{Name: "online_access_info_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"shopify_user_id", "first_name", "last_name", "email", "account_owner", "locale", "collaborator", "email_verified", "updated_at"}),
			}).Create(&associatedUser).Error; err != nil {
				s.Log.Errorf("Failed to store associated user: %+v", err)
				return err
			}

			// return nil will commit the whole transaction
			return nil
		})

		if err != nil {
			s.Log.Errorf("Failed to store shopify token: %+v", err)
			return nil, err
		}
		return &user, nil
	}

	return nil, fiber.NewError(fiber.StatusBadRequest, "Online access info not found!")
}
