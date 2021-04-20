package models

import "database/sql"

type Role struct {
	RoleID                   int             `db:"role_id" json:"role_id"`
	RoleName                 string          `db:"name" json:"role_name"`
	RoleDescription          *sql.NullString `db:"description" json:"-"`
	RoleSpecifiedDescription string          `json:"role_description"`
	AllowRolesCrud           bool            `db:"allow_roles_crud" json:"allow_roles_crud"`
	AllowUsersCrud           bool            `db:"allow_users_crud" json:"allow_users_crud"`
	AllowVeterinariansCrud   bool            `db:"allow_veterinarians_crud" json:"allow_veterinarians_crud"`
	AllowVaccinesCrud        bool            `db:"allow_vaccines_crud" json:"allow_vaccines_crud"`
	AllowFoodCrud            bool            `db:"allow_food_crud" json:"allow_food_crud"`
	AllowPetsCrud            bool            `db:"allow_pets_crud" json:"allow_pets_crud"`
	AllowDatabaseDump        bool            `db:"allow_database_dump" json:"allow_database_dump"`
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
	if r.RoleDescription.Valid {
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
