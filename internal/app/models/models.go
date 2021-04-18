package models

import "database/sql"

type User struct {
	UserID         int             `db:"user_id"`
	AccountEmail   string          `db:"account_email"`
	PasswordSHA256 string          `db:"password_sha256"`
	Username       string          `db:"username"`
	FullName       string          `db:"full_name"`
	BackupEmail    *sql.NullString `db:"backup_email"`
	Location       *sql.NullString `db:"location"`
}
