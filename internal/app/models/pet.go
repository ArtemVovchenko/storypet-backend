package models

import (
	"database/sql"
	validation "github.com/go-ozzo/ozzo-validation"
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

func (p *PetType) Update(other *PetType) {
	if other.RERCoefficient != 0 && other.RERCoefficient != p.RERCoefficient {
		p.RERCoefficient = other.RERCoefficient
	}
	if other.TypeName != "" && other.TypeName != p.TypeName {
		p.TypeName = other.TypeName
	}
}

type Pet struct {
	PetID          int             `json:"pet_id" db:"pet_id"`
	Name           string          `json:"name" db:"name"`
	UserID         int             `json:"user_id" db:"user_id"`
	PetType        int             `json:"pet_type" db:"pet_type"`
	MotherVerified bool            `json:"mother_verified" db:"mother_verified"`
	FatherVerified bool            `json:"father_verified" db:"father_verified"`
	MotherID       *sql.NullInt64  `json:"-" db:"mother_id"`
	FatherID       *sql.NullInt64  `json:"-" db:"father_id"`
	VeterinarianID *sql.NullInt64  `json:"-" db:"veterinarian_id"`
	Breed          *sql.NullString `json:"-" db:"breed"`
	FamilyName     *sql.NullString `json:"-" db:"family_name"`

	SpecifiedMotherID       int    `json:"mother_id,omitempty"`
	SpecifiedFatherID       int    `json:"father_id,omitempty"`
	SpecifiedVeterinarianID int    `json:"veterinarian_id,omitempty"`
	SpecifiedBreed          string `json:"breed,omitempty"`
	SpecifiedFamilyName     string `json:"family_name,omitempty"`
}

func (p *Pet) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.Name, validation.Required, validation.Length(2, 30)),
		validation.Field(&p.PetType, validation.Required))
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
	if p.SpecifiedBreed != "" {
		p.Breed = &sql.NullString{
			String: p.SpecifiedBreed,
			Valid:  true,
		}
	}
	if p.SpecifiedFamilyName != "" {
		p.FamilyName = &sql.NullString{
			String: p.SpecifiedFamilyName,
			Valid:  true,
		}
	}

}

func (p *Pet) AfterCreate() {
	if p.Breed != nil && p.Breed.Valid {
		p.SpecifiedBreed = p.Breed.String
	}
	if p.FamilyName != nil && p.FamilyName.Valid {
		p.SpecifiedFamilyName = p.FamilyName.String
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

func (p *Pet) Update(other *Pet) {
	// TODO: Model provides a part of business logic (verification of parents). Fix it after LB
	if other.PetType != 0 && p.PetType != other.PetType {
		p.PetType = other.PetType
	}
	if other.Name != "" && p.Name != other.Name {
		p.Name = other.Name
	}
	if other.UserID != 0 && p.UserID != other.UserID {
		p.UserID = other.UserID
	}
	if other.MotherVerified != p.MotherVerified {
		p.MotherVerified = other.MotherVerified
	}
	if other.FatherVerified != p.FatherVerified {
		p.FatherVerified = other.FatherVerified
	}
	if other.SpecifiedMotherID != p.SpecifiedMotherID {
		if other.SpecifiedMotherID == 0 {
			p.SpecifiedMotherID = 0
			p.MotherID = nil
		} else {
			p.SpecifiedMotherID = other.SpecifiedMotherID
			p.MotherVerified = false
		}
	}
	if other.SpecifiedFatherID != p.SpecifiedFatherID {
		if other.SpecifiedFatherID == 0 {
			p.SpecifiedFatherID = 0
			p.FatherID = nil
		} else {
			p.SpecifiedFatherID = other.SpecifiedFatherID
			p.FatherVerified = false
		}
	}
	if other.SpecifiedVeterinarianID != p.SpecifiedVeterinarianID {
		if other.SpecifiedVeterinarianID == 0 {
			p.SpecifiedVeterinarianID = 0
			p.VeterinarianID = nil
		} else {
			p.SpecifiedVeterinarianID = other.SpecifiedVeterinarianID
		}
	}
	if other.SpecifiedBreed != p.SpecifiedBreed {
		if other.SpecifiedBreed == "" {
			p.Breed = nil
		}
		p.SpecifiedBreed = other.SpecifiedBreed
	}
	if other.SpecifiedFamilyName != p.SpecifiedFamilyName {
		if other.SpecifiedFamilyName == "" {
			p.FamilyName = nil
		}
		p.SpecifiedFamilyName = other.SpecifiedFamilyName
	}
}

func (p *Pet) SetSpecifiedMotherID(id *int) {
	if id != nil {
		p.SpecifiedMotherID = *id
	}
}

func (p *Pet) SetSpecifiedFatherID(id *int) {
	if id != nil {
		p.SpecifiedFatherID = *id
	}
}

func (p *Pet) SetSpecifiedVeterinarianID(id *int) {
	if id != nil {
		p.SpecifiedVeterinarianID = *id
	}
}

func (p *Pet) SetSpecifiedBreed(breed *string) {
	if breed != nil {
		p.SpecifiedBreed = *breed
	}
}

func (p *Pet) SetSpecifiedFamilyName(name *string) {
	if name != nil {
		p.SpecifiedFamilyName = *name
	}
}

type Anthropometry struct {
	RecordID int       `json:"record_id" db:"record_id"`
	PetID    int       `json:"pet_id" db:"pet_id"`
	Time     time.Time `json:"record_time" db:"record_time"`
	Height   float64   `json:"height" db:"height"`
	Weight   float64   `json:"weight" db:"weight"`
}

func (a *Anthropometry) Validate() error {
	return validation.ValidateStruct(
		a,
		validation.Field(&a.Height, validation.Required, validation.Min(0.01)),
		validation.Field(&a.Weight, validation.Required, validation.Min(0.01)),
	)
}

type Activity struct {
	PetID           int       `json:"pet_id" db:"pet_id"`
	RecordTimestamp time.Time `json:"record_timestamp" db:"record_timestamp"`
	Distance        float64   `json:"distance" db:"distance"`
	MeanSpeed       float64   `json:"mean_speed" db:"mean_speed"`
}

type PetHealthReport struct {
	PetID                   int             `json:"pet_id" db:"pet_id"`
	ReportTimestamp         time.Time       `json:"report_timestamp" db:"report_timestamp"`
	VeterinarianID          int             `json:"creator_id" db:"veterinarian_id"`
	ReportConclusion        string          `json:"report_conclusion" db:"report_conclusion"`
	ReportComments          *sql.NullString `json:"-" db:"report_comments"`
	SpecifiedReportComments string          `json:"report_comment"`
}

func (p *PetHealthReport) BeforeCreate() {
	if p.SpecifiedReportComments != "" {
		p.ReportComments = &sql.NullString{
			String: p.SpecifiedReportComments,
			Valid:  true,
		}
	}
}

func (p *PetHealthReport) SetSpecifiedReportComments(reportComments *string) {
	if reportComments != nil {
		p.SpecifiedReportComments = *reportComments
	}
}

func (p *PetHealthReport) AfterCreate() {
	if p.ReportComments != nil && p.ReportComments.Valid {
		p.SpecifiedReportComments = p.ReportComments.String
	}
}

type FoodCaloriesReport struct {
	Date              time.Time `db:"date" json:"date"`
	FoodTotalCalories float64   `db:"eat_ccal" json:"food_total_calories"`
}

type RERCaloriesReport struct {
	Date             time.Time `db:"date" json:"date"`
	RERTotalCalories float64   `db:"rer_ccal" json:"rer_total_calories"`
}

type AnthropometryReport struct {
	Date   time.Time `json:"date" db:"date"`
	Weight float64   `json:"weight" db:"weight"`
	Height float64   `json:"height" db:"height"`
}

type TodayReport struct {
	FoodTotalCalories float64 `db:"eat_ccal" json:"food_total_calories"`
	RERTotalCalories  float64 `db:"rer_ccal" json:"rer_total_calories"`
}
