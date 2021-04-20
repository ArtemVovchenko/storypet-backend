package models

import (
	"database/sql"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID               int             `db:"user_id" json:"user_id"`
	AccountEmail         string          `db:"account_email" json:"account_email"`
	PasswordSHA256       string          `db:"password_sha256" json:"-"`
	Username             string          `db:"username" json:"username"`
	FullName             string          `db:"full_name" json:"full_name"`
	BackupEmail          *sql.NullString `db:"backup_email" json:"-"`
	Location             *sql.NullString `db:"location" json:"-"`
	Password             string          `json:"-"`
	SpecifiedBackupEmail string          `json:"backup_email,omitempty"`
	SpecifiedLocation    string          `json:"location,omitempty"`
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
		validation.Field(&u.SpecifiedBackupEmail, is.Email),
	)
}

func encryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
