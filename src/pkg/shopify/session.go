package shopify

import (
	"errors"
	"fmt"
	"time"

	"github.com/mbbipul/prodigyai-backend/src/pkg/shopify/utils"
)

type AssociatedUser struct {
	ID            int64  `json:"id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	AccountOwner  bool   `json:"account_owner"`
	Locale        string `json:"locale"`
	Collaborator  bool   `json:"collaborator"`
}

type OnlineAccessInfo struct {
	ExpiresIn           int            `json:"expires_in" validate:"required"`
	AssociatedUserScope string         `json:"associated_user_scope" validate:"required"`
	AssociatedUser      AssociatedUser `json:"associated_user" validate:"required"`
}

type Session struct {
	ID               string            `json:"id" validate:"required"`
	Shop             string            `json:"shop" validate:"required"`
	State            string            `json:"state" validate:"required"`
	IsOnline         bool              `json:"isOnline" validate:"required"`
	Scope            *string           `json:"scope"`
	Expires          time.Time         `json:"expires"`
	AccessToken      *string           `json:"accessToken"`
	OnlineAccessInfo *OnlineAccessInfo `json:"onlineAccessInfo"`
}

func GetSessionExpiration(expiresIn int) time.Time {
	return time.Now().Add(time.Duration(expiresIn) * time.Second)
}

func GetSessionID(shop string, userId int64) (string, error) {
	sanitizedShop, err := utils.SanitizeShop(shop, true)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s_%d", sanitizedShop, userId), nil
}

func CreateSession(exchangeResponse *TokenExchangeResponse, shop string) (*Session, error) {
	if exchangeResponse == nil {
		return nil, errors.New("exchange response is nil")
	}
	sessionId, err := GetSessionID(shop, exchangeResponse.AssociatedUser.ID)
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:          sessionId,
		Shop:        shop,
		AccessToken: &exchangeResponse.AccessToken,
		Expires:     GetSessionExpiration(exchangeResponse.ExpiresIn),
		IsOnline:    true,
		Scope:       &exchangeResponse.Scope,
		State:       "active",
		OnlineAccessInfo: &OnlineAccessInfo{
			ExpiresIn:           exchangeResponse.ExpiresIn,
			AssociatedUserScope: exchangeResponse.AssociatedUserScope,
			AssociatedUser:      exchangeResponse.AssociatedUser,
		},
	}

	return session, nil
}
