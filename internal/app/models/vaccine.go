package models

import (
	"database/sql"
	validation "github.com/go-ozzo/ozzo-validation"
	"time"
)

type Vaccine struct {
	VaccineID            int             `json:"vaccine_id" db:"vaccine_id"`
	PetID                int             `json:"pet_id" db:"pet_id"`
	Name                 string          `json:"name" db:"name"`
	VaccinationDate      time.Time       `json:"vaccination_date" db:"vaccination_date"`
	Description          *sql.NullString `json:"-" db:"description"`
	SpecifiedDescription string          `json:"specified_description,omitempty"`
}

func (v *Vaccine) BeforeCreate() {
	if v.SpecifiedDescription != "" {
		v.Description = &sql.NullString{
			String: v.SpecifiedDescription,
			Valid:  true,
		}
	}
}

func (v *Vaccine) AfterCreate() {
	if v.Description != nil && v.Description.Valid {
		v.SpecifiedDescription = v.Description.String
	}
}

func (v *Vaccine) Update(other *Vaccine) {
	v.AfterCreate()
	if other.PetID != v.PetID {
		v.PetID = other.PetID
	}
	if other.VaccinationDate != v.VaccinationDate {
		v.VaccinationDate = other.VaccinationDate
	}
	if other.SpecifiedDescription != v.SpecifiedDescription {
		v.SpecifiedDescription = other.SpecifiedDescription
	}
	v.BeforeCreate()
}

func (v *Vaccine) Validate() error {
	return validation.ValidateStruct(
		v,
		validation.Field(&v.PetID, validation.Required),
		validation.Field(&v.Name, validation.Required, validation.Length(3, 20)),
		validation.Field(&v.VaccinationDate, validation.Required),
		validation.Field(&v.SpecifiedDescription, validation.Length(3, 0)),
	)

}

func (v *Vaccine) SetSpecifiedDescription(specifiedDescription *string) {
	if specifiedDescription != nil {
		v.SpecifiedDescription = *specifiedDescription
	}
}
