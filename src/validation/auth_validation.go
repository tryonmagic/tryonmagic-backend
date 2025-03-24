package validation

import (
	"app/src/model"
	"time"
)

type Register struct {
	FirstName string     `json:"firstName" validate:"required,max=50" example:"fake name"`
	LastName  string     `json:"lastName" validate:"required,max=50" example:"fake name"`
	Email     string     `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
	Password  string     `json:"password" validate:"required,min=8,max=20,password" example:"password1"`
	Role      model.Role `json:"role" validate:"required,oneof=USER ADMIN INTERNAL_SERVICE,max=50" example:"USER"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
	Password string `json:"password" validate:"required,min=8,max=20,password" example:"password1"`
}

type GoogleLogin struct {
	FirstName     string `json:"firstName" validate:"required,max=50"`
	LastName      string `json:"lastName" validate:"required,max=50"`
	Email         string `json:"email" validate:"required,email,max=50"`
	EmailVerified bool   `json:"emailVerified" validate:"required"`
}

type Logout struct {
	RefreshToken string `json:"refresh_token" validate:"required,max=255"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token" validate:"required,max=255"`
}

type ForgotPassword struct {
	Email string `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
}

type Token struct {
	Token string `json:"token" validate:"required,max=255"`
}

type OnlineAccessUser struct {
	ID            int64  `json:"id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	AccountOwner  bool   `json:"account_owner"`
	Locale        string `json:"locale"`
	Collaborator  bool   `json:"collaborator"`
}

type ShopifyOnlineAccessInfo struct {
	ExpiresIn           int              `json:"expires_in" validate:"required"`
	AssociatedUserScope string           `json:"associated_user_scope" validate:"required"`
	AssociatedUser      OnlineAccessUser `json:"associated_user" validate:"required"`
}

type StoreShopifyToken struct {
	ID               string                   `json:"id" validate:"required"`
	Shop             string                   `json:"shop" validate:"required"`
	State            string                   `json:"state" validate:"required"`
	IsOnline         bool                     `json:"isOnline" validate:"required"`
	Scope            *string                  `json:"scope"`
	Expires          *time.Time               `json:"expires"`
	AccessToken      *string                  `json:"accessToken"`
	OnlineAccessInfo *ShopifyOnlineAccessInfo `json:"onlineAccessInfo"`
}
