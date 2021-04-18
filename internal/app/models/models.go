package models

type User struct {
	UserID         int    `db:"user_id"`
	AccountEmail   string `db:"account_email"`
	PasswordSHA256 string `db:"password_sha256"`
	Username       string `db:"username"`
	FullName       string `db:"full_name"`
	BackupEmail    string `db:"backup_email"`
	Location       string `db:"location"`
}
