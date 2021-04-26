package models

import (
	"database/sql"
	"reflect"
)

type Role struct {
	RoleID                   int             `db:"role_id" json:"role_id"`
	RoleName                 string          `db:"name" json:"role_name"`
	RoleDescription          *sql.NullString `db:"description" json:"-"`
	RoleSpecifiedDescription string          `json:"role_description,omitempty"`
	AllowRolesCrud           bool            `db:"allow_roles_crud" json:"allow_roles_crud"`
	AllowUsersCrud           bool            `db:"allow_users_crud" json:"allow_users_crud"`
	AllowVeterinariansCrud   bool            `db:"allow_veterinarians_crud" json:"allow_veterinarians_crud"`
	AllowVaccinesCrud        bool            `db:"allow_vaccines_crud" json:"allow_vaccines_crud"`
	AllowFoodCrud            bool            `db:"allow_food_crud" json:"allow_food_crud"`
	AllowPetsCrud            bool            `db:"allow_pets_crud" json:"allow_pets_crud"`
	AllowDatabaseDump        bool            `db:"allow_database_dump" json:"allow_database_dump"`
}

type Permission struct {
	Name string
}

// Permissions MUST HAVE all the bool permission fields of
// models.Role structure.
//
//Any change of Permissions set must be updated
// in permissions package
type Permissions struct {
	RolesPermission         Permission
	UsersPermission         Permission
	VeterinariansPermission Permission
	VaccinesPermission      Permission
	FoodPermission          Permission
	PetsPermission          Permission
	DatabasePermission      Permission
}

func (r *Role) BeforeCreate() {
	if r.RoleSpecifiedDescription != "" {
		r.RoleDescription = &sql.NullString{
			String: r.RoleSpecifiedDescription,
			Valid:  true,
		}
	} else {
		r.RoleDescription = &sql.NullString{
			String: "",
			Valid:  false,
		}
	}
}

func (r *Role) CheckNullableData() {
	if r.RoleDescription != nil && r.RoleDescription.Valid {
		r.RoleSpecifiedDescription = r.RoleDescription.String
	}
}

func (r *Role) SetDescription(description *string) {
	if description != nil {
		r.RoleSpecifiedDescription = *description
	} else {
		r.RoleSpecifiedDescription = ""
	}
}

func (r *Role) Update(other *Role) {
	r.RoleName = other.RoleName
	r.RoleSpecifiedDescription = other.RoleSpecifiedDescription
	r.AllowRolesCrud = other.AllowRolesCrud
	r.AllowUsersCrud = other.AllowUsersCrud
	r.AllowVeterinariansCrud = other.AllowVeterinariansCrud
	r.AllowVaccinesCrud = other.AllowVaccinesCrud
	r.AllowFoodCrud = other.AllowFoodCrud
	r.AllowPetsCrud = other.AllowPetsCrud
	r.AllowDatabaseDump = other.AllowDatabaseDump

	r.BeforeCreate()
}

func (r Role) HasAllPermission(permissions ...Permission) bool {
	for _, perm := range permissions {
		if !reflect.ValueOf(r).FieldByName(perm.Name).Bool() {
			return false
		}
	}
	return true
}

func (r Role) HasAnyPermission(permissions ...Permission) bool {
	for _, perm := range permissions {
		if reflect.ValueOf(r).FieldByName(perm.Name).Bool() {
			return true
		}
	}
	return false
}

func RolePermissions() *Permissions {
	return &Permissions{
		RolesPermission:         Permission{"AllowRolesCrud"},
		UsersPermission:         Permission{"AllowUsersCrud"},
		VeterinariansPermission: Permission{"AllowVeterinariansCrud"},
		VaccinesPermission:      Permission{"AllowVaccinesCrud"},
		FoodPermission:          Permission{"AllowFoodCrud"},
		PetsPermission:          Permission{"AllowPetsCrud"},
		DatabasePermission:      Permission{"AllowDatabaseDump"},
	}
}
