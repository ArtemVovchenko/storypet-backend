package store

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/configs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepository_FindByAccountEmail(t *testing.T) {
	store := New(configs.NewDatabaseConfig())
	if err := store.Open(); err != nil {
		return
	}
	defer store.Close(nil)
	expected := models.User{
		UserID:         1,
		AccountEmail:   "vovchenko.artem@icloud.com",
		PasswordSHA256: "b68e832ce1dbec1d37f79d5ea28ab0cb628d8339a27261336208786825ce9826",
		Username:       "an_unseen_future",
		FullName:       "Artem Vovchenko",
		BackupEmail:    "temka.vovchenko@gmail.com",
		Location:       "Kharkiv, Ukraine",
	}

	actual, err := store.userRepository.FindByAccountEmail("vovchenko.artem@ocloud.com")
	if err != nil {
		actual = nil
	}
	assert.Equal(t, expected, actual)
}

func TestUserRepository_FindByAccountEmail2(t *testing.T) {
	store := New(configs.NewDatabaseConfig())
	if err := store.Open(); err != nil {
		return
	}
	defer store.Close(nil)
	expected := models.User{
		UserID:         2,
		AccountEmail:   "teresi.alex.@icloud.com",
		PasswordSHA256: "b68e832ce1dbec1d37f79d5ea28ab0cb628d8339a27261336208786825ce9826",
		Username:       "teresin",
		FullName:       "Teresin Alexandr",
		BackupEmail:    "terx.a@i.ua",
		Location:       "Kiyiv, Ukraine",
	}

	actual, err := store.userRepository.FindByAccountEmail("vovchenko.artem@ocloud.com")
	if err != nil {
		actual = nil
	}
	assert.Equal(t, expected, actual)
}
