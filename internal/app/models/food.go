package models

import (
	"database/sql"
	validation "github.com/go-ozzo/ozzo-validation"
)

type Food struct {
	FoodID                int             `json:"food_id" db:"food_id"`
	FoodName              string          `json:"food_name" db:"food_name"`
	Calories              float64         `json:"food_calories" db:"calories"`
	Description           *sql.NullString `json:"-" db:"description"`
	Manufacturer          *sql.NullString `json:"-" db:"manufacturer"`
	CreatorID             *sql.NullInt64  `json:"-" db:"creator_id"`
	SpecifiedDescription  string          `json:"description,omitempty"`
	SpecifiedManufacturer string          `json:"manufacturer,omitempty"`
	SpecifiedCreatorID    int             `json:"creator_id"`
}

func (f *Food) Validate() error {
	return validation.ValidateStruct(
		f,
		validation.Field(&f.FoodName, validation.Required, validation.Length(3, 255)),
		validation.Field(&f.Calories, validation.Required, validation.Min(0.0)),
		validation.Field(&f.SpecifiedDescription, validation.Length(5, 255)),
		validation.Field(&f.SpecifiedManufacturer, validation.Length(2, 255)),
	)
}

func (f *Food) BeforeCreate() {
	if f.SpecifiedDescription != "" {
		f.Description = &sql.NullString{
			String: f.SpecifiedDescription,
			Valid:  true,
		}
	}
	if f.SpecifiedManufacturer != "" {
		f.Manufacturer = &sql.NullString{
			String: f.SpecifiedManufacturer,
			Valid:  true,
		}
	}
	if f.SpecifiedCreatorID != 0 {
		f.CreatorID = &sql.NullInt64{
			Int64: int64(f.SpecifiedCreatorID),
			Valid: true,
		}
	}
}

func (f *Food) AfterCreate() {
	if f.Description != nil && f.Description.Valid {
		f.SpecifiedDescription = f.Description.String
	}
	if f.Manufacturer != nil && f.Manufacturer.Valid {
		f.SpecifiedManufacturer = f.Manufacturer.String
	}
	if f.CreatorID != nil && f.CreatorID.Valid {
		f.SpecifiedCreatorID = int(f.CreatorID.Int64)
	}
}

func (f *Food) SetSpecifiedDescription(description *string) {
	if description != nil {
		f.SpecifiedDescription = *description
	}
}
func (f *Food) SetSpecifiedManufacturer(manufacturer *string) {
	if manufacturer != nil {
		f.SpecifiedManufacturer = *manufacturer
	}
}
