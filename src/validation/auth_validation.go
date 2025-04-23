package validation

import (
	"github.com/mbbipul/prodigyai-backend/src/model"
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
	RefreshToken string `json:"refreshToken" validate:"required,max=255"`
}

type RefreshToken struct {
	RefreshToken string `json:"refreshToken" validate:"required,max=255"`
}

type ForgotPassword struct {
	Email string `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
}

type Token struct {
	Token string `json:"token" validate:"required,max=255"`
}

type ShopifyAuth struct {
	SessionToken string `json:"sessionToken" validate:"required"`
}
