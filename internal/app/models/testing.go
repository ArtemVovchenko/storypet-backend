package models

import (
	"testing"
)

func TestUser(t *testing.T) *User {
	u := &User{
		AccountEmail:         "blabla@gmail.com",
		Password:             "123123",
		Username:             "blashka",
		FullName:             "Yurii Ivanitskiy",
		SpecifiedBackupEmail: "yi@gmail.com",
		SpecifiedLocation:    "Kiyiv, Ukraine",
	}
	return u
}
