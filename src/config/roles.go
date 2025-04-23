package config

import (
	// initialize the Swagger documentation
	"github.com/mbbipul/prodigyai-backend/src/model"
)

var allRoles = map[model.Role][]string{
	model.RoleUser:            {},
	model.RoleAdmin:           {"getUsers", "manageUsers"},
	model.RoleInternalService: {},
}

var Roles = getKeys(allRoles)
var RoleRights = allRoles

func getKeys(m map[model.Role][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, string(k))
	}
	return keys
}
