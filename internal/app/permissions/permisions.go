package permissions

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

// All returns all the permissions, registered in system
func All() []models.Permission {
	return []models.Permission{
		{"AllowRolesCrud"},
		{"AllowUsersCrud"},
		{"AllowVeterinariansCrud"},
		{"AllowVaccinesCrud"},
		{"AllowFoodCrud"},
		{"AllowPetsCrud"},
		{"AllowDatabaseDump"},
	}
}

func Permissions() *models.Permissions {
	return models.RolePermissions()
}

func AllRolesHavePermissions(roles []models.Role, permissions ...models.Permission) bool {
	for _, role := range roles {
		if !role.HasAllPermission(permissions...) {
			return false
		}
	}
	return true
}

func AnyRoleHavePermissions(roles []models.Role, permissions ...models.Permission) bool {
	for _, role := range roles {
		if role.HasAllPermission(permissions...) {
			return true
		}
	}
	return false
}

func AnyRoleIsVeterinarian(roles []models.Role) bool {
	for _, role := range roles {
		if role.IsVeterinarian() {
			return true
		}
	}
	return false
}
