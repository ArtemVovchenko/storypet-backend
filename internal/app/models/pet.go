package models

import (
	"database/sql"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"time"
)

type PetType struct {
	TypeID         int     `json:"type_id" db:"type_id"`
	TypeName       string  `json:"type_name" db:"type_name"`
	RERCoefficient float64 `json:"rer_coefficient" db:"rer_coefficient"`
}

func (p *PetType) Validate() error {
	return validation.ValidateStruct(
		p,
		validation.Field(&p.TypeName, validation.Required),
		validation.Field(&p.RERCoefficient, validation.Required),
	)
}

type Pet struct {
	PetID          int             `json:"pet_id" db:"pet_id"`
	Name           int             `json:"name" db:"name"`
	MotherVerified bool            `json:"mother_verified" db:"mother_verified"`
	FatherVerified bool            `json:"father_verified" db:"father_verified"`
	MotherID       *sql.NullInt64  `json:"-" db:"mother_id"`
	FatherID       *sql.NullInt64  `json:"-" db:"father_id"`
	VeterinarianID *sql.NullInt64  `json:"-" db:"veterinarian_id"`
	Breed          *sql.NullString `json:"-" db:"breed"`
	FamilyName     *sql.NullString `json:"-" db:"family_name"`

	SpecifiedMotherID       int     `json:"mother_id,omitempty"`
	SpecifiedFatherID       int     `json:"father_id,omitempty"`
	SpecifiedVeterinarianID int     `json:"veterinarian_id,omitempty"`
	SpecifiedBreed          *string `json:"breed,omitempty"`
	SpecifiedFamilyName     *string `json:"family_name,omitempty"`
}

func (p *Pet) Validate() error {
	return validation.ValidateStruct(p, validation.Field(&p.Name, validation.Required, is.Alpha))
}

func (p *Pet) BeforeCreate() {
	if p.SpecifiedFatherID != 0 {
		p.FatherID = &sql.NullInt64{
			Int64: int64(p.SpecifiedFatherID),
			Valid: true,
		}
	}
	if p.SpecifiedMotherID != 0 {
		p.MotherID = &sql.NullInt64{
			Int64: int64(p.SpecifiedMotherID),
			Valid: true,
		}
	}
	if p.SpecifiedVeterinarianID != 0 {
		p.VeterinarianID = &sql.NullInt64{
			Int64: int64(p.SpecifiedVeterinarianID),
			Valid: true,
		}
	}
	if p.SpecifiedBreed != nil {
		p.Breed = &sql.NullString{
			String: *p.SpecifiedBreed,
			Valid:  true,
		}
	}
	if p.SpecifiedFamilyName != nil {
		p.FamilyName = &sql.NullString{
			String: *p.SpecifiedFamilyName,
			Valid:  true,
		}
	}

}

func (p *Pet) AfterCreate() {
	if p.Breed != nil && p.Breed.Valid {
		p.SpecifiedBreed = &p.Breed.String
	}
	if p.FamilyName != nil && p.FamilyName.Valid {
		p.SpecifiedFamilyName = &p.FamilyName.String
	}
	if p.VeterinarianID != nil && p.VeterinarianID.Valid {
		p.SpecifiedVeterinarianID = int(p.VeterinarianID.Int64)
	}
	if p.MotherID != nil && p.MotherID.Valid {
		p.SpecifiedMotherID = int(p.MotherID.Int64)
	}
	if p.FatherID != nil && p.FatherID.Valid {
		p.SpecifiedFatherID = int(p.FatherID.Int64)
	}
}

type Anthropometry struct {
	PetID  int       `json:"pet_id" db:"pet_id"`
	Time   time.Time `json:"time" db:"time"`
	Height float64   `json:"height" db:"height"`
	Weight float64   `json:"weight" db:"weight"`
}

type Activity struct {
	PetID           int       `json:"pet_id" db:"pet_id"`
	RecordTimestamp time.Time `json:"record_timestamp" db:"record_timestamp"`
	Distance        float64   `json:"distance" db:"distance"`
	PeakSpeed       float64   `json:"peak_speed" db:"peak_speed"`
}

type PetHealthReport struct {
	PetID                   int             `json:"pet_id" db:"pet_id"`
	ReportTimestamp         time.Time       `json:"report_timestamp" db:"report_timestamp"`
	VeterinarianID          int             `json:"veterinarian_id" db:"veterinarian_id"`
	ReportConclusion        string          `json:"report_conclusion" db:"report_conclusion"`
	ReportComments          *sql.NullString `json:"-" db:"report_comments"`
	SpecifiedReportComments string          `json:"report_comments"`
}

func (p *PetHealthReport) BeforeCreate() {
	if p.SpecifiedReportComments != "" {
		p.ReportComments = &sql.NullString{
			String: p.SpecifiedReportComments,
			Valid:  true,
		}
	}
}

func (p *PetHealthReport) AfterCreate() {
	if p.ReportComments != nil && p.ReportComments.Valid {
		p.SpecifiedReportComments = p.ReportComments.String
	}
}
