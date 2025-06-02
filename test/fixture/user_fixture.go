package fixture

import (
	"github.com/mbbipul/tryonmagic-backend/src/model"

	"github.com/google/uuid"
)

var UserOne = &model.User{
	ID:            uuid.New(),
	FirstName:     "Test1",
	Email:         "test1@gmail.com",
	Password:      "password1",
	Role:          "user",
	EmailVerified: false,
}

var UserTwo = &model.User{
	ID:            uuid.New(),
	FirstName:     "Test2",
	Email:         "test2@gmail.com",
	Password:      "password1",
	Role:          "user",
	EmailVerified: false,
}

var Admin = &model.User{
	ID:            uuid.New(),
	FirstName:     "Admin",
	Email:         "admin@gmail.com",
	Password:      "password1",
	Role:          "admin",
	EmailVerified: false,
}
