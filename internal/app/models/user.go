package models

import (
	"database/sql"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	UserID               int             `db:"user_id" json:"user_id"`
	AccountEmail         string          `db:"account_email" json:"account_email"`
	PasswordSHA256       string          `db:"password_sha256" json:"-"`
	Username             string          `db:"username" json:"username"`
	FullName             string          `db:"full_name" json:"full_name"`
	RegistrationDate     *time.Time      `db:"registration_date" json:"registration_date"`
	SubscriptionDate     *time.Time      `db:"subscription_date" json:"subscription_date"`
	BackupEmail          *sql.NullString `db:"backup_email" json:"-"`
	Location             *sql.NullString `db:"location" json:"-"`
	Password             string          `json:"-"`
	SpecifiedBackupEmail string          `json:"backup_email,omitempty"`
	SpecifiedLocation    string          `json:"location,omitempty"`
}

type VetClinic struct {
	UserID     int    `json:"user_id" db:"user_id"`
	ClinicID   string `json:"clinic_id" db:"clinic_id"`
	ClinicName string `json:"clinic_name" db:"clinic_name"`
}

func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := encryptString(u.Password)
		if err != nil {
			return err
		}
		u.PasswordSHA256 = enc
	}

	if u.SpecifiedLocation != "" {
		u.Location = &sql.NullString{
			String: u.SpecifiedLocation,
			Valid:  true,
		}
	} else {
		u.Location = &sql.NullString{
			String: u.SpecifiedLocation,
			Valid:  false,
		}
	}

	if u.SpecifiedBackupEmail != "" {
		u.BackupEmail = &sql.NullString{
			String: u.SpecifiedBackupEmail,
			Valid:  true,
		}
	} else {
		u.BackupEmail = &sql.NullString{
			String: u.SpecifiedBackupEmail,
			Valid:  false,
		}
	}

	return nil
}

func (u *User) AfterCreate() {
	if u.BackupEmail != nil && u.BackupEmail.Valid {
		u.SpecifiedBackupEmail = u.BackupEmail.String
	}
	if u.Location != nil && u.Location.Valid {
		u.SpecifiedLocation = u.Location.String
	}
}

func (u *User) SetLocation(location *string) {
	if location != nil {
		u.SpecifiedLocation = *location
	} else {
		u.SpecifiedLocation = ""
	}
}

func (u *User) SetBackupEmail(backupEmail *string) {
	if backupEmail != nil {
		u.SpecifiedBackupEmail = *backupEmail
	} else {
		u.SpecifiedBackupEmail = ""
	}
}

func (u *User) ComparePasswords(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordSHA256), []byte(password)) == nil
}

func (u *User) Sanitise() {
	if u.BackupEmail.Valid {
		u.SpecifiedBackupEmail = u.BackupEmail.String
	} else {
		u.SpecifiedBackupEmail = ""
	}

	if u.Location.Valid {
		u.SpecifiedLocation = u.Location.String
	} else {
		u.SpecifiedLocation = ""
	}
}

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.AccountEmail, validation.Required, is.Email),
		validation.Field(&u.Password, validation.By(requiredIf(u.PasswordSHA256 != "")), is.Alphanumeric, validation.Length(6, 100)),
		validation.Field(&u.Username, validation.Required, validation.Length(5, 30)),
		validation.Field(&u.FullName, validation.Required, validation.Length(5, 30)),
	)
}

func (u *User) Update(other *User) {
	if u.AccountEmail != other.AccountEmail && len(other.AccountEmail) > 0 {
		u.AccountEmail = other.AccountEmail
	}
	if u.FullName != other.FullName && len(other.FullName) > 0 {
		u.FullName = other.FullName
	}
	if u.Username != other.Username && len(other.Username) > 0 {
		u.Username = other.Username
	}
	if u.SpecifiedLocation != other.SpecifiedLocation && len(other.SpecifiedLocation) > 0 {
		u.SpecifiedLocation = other.SpecifiedLocation
	}
	if u.SpecifiedBackupEmail != other.SpecifiedBackupEmail && len(other.SpecifiedBackupEmail) > 0 {
		u.SpecifiedBackupEmail = other.SpecifiedBackupEmail
	}
}

func encryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
