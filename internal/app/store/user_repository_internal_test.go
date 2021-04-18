package store

import (
	"database/sql"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/configs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepository_FindByAccountEmail(t *testing.T) {
	store := New(configs.NewDatabaseConfig())
	_ = store.Open()
	defer store.Close()
	expected := &models.User{
		UserID:         1,
		AccountEmail:   "vovchenko.artem@icloud.com",
		PasswordSHA256: "b68e832ce1dbec1d37f79d5ea28ab0cb628d8339a27261336208786825ce9826",
		Username:       "an_unseen_future",
		FullName:       "Artem Vovchenko",
		BackupEmail:    &sql.NullString{String: "temka.vovchenko@gmail.com"},
		Location:       &sql.NullString{String: "Kharkiv, Ukraine"},
	}

	actual, err := store.Users().FindByAccountEmail("vovchenko.artem@icloud.com")
	assert.Equal(t, nil, err)

	assert.Equal(t, expected.AccountEmail, actual.AccountEmail)
	assert.Equal(t, expected.PasswordSHA256, actual.PasswordSHA256)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.FullName, actual.FullName)
	assert.Equal(t, expected.BackupEmail.String, actual.BackupEmail.String)
	assert.Equal(t, expected.Location.String, actual.Location.String)
}

func TestUserRepository_FindByAccountEmail2(t *testing.T) {
	store := New(configs.NewDatabaseConfig())
	_ = store.Open()
	defer store.Close()
	expected, _ := store.Users().FindByAccountEmail("vovchenko.artem@icloud.com")
	actual, _ := store.Users().FindByAccountEmail("vovchenko.artem@ocloud.com")
	assert.NotEqual(t, expected, actual)
}

func TestUserRepository_Create(t *testing.T) {
	store := New(configs.NewDatabaseConfig())
	_ = store.Open()
	defer store.Close()

	expected := &models.User{
		UserID:         0,
		AccountEmail:   "blabla@gmail.com",
		PasswordSHA256: "123",
		Username:       "blashka",
		FullName:       "Yurii Ivanitskiy",
		BackupEmail:    &sql.NullString{String: "yi@gmail.com", Valid: true},
		Location:       &sql.NullString{String: "Kiyiv, Ukraine", Valid: true},
	}

	actual, err := store.Users().Create(expected)
	assert.Equal(t, nil, err)

	assert.Equal(t, expected.AccountEmail, actual.AccountEmail)
	assert.Equal(t, expected.PasswordSHA256, actual.PasswordSHA256)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.FullName, actual.FullName)
	assert.Equal(t, expected.BackupEmail.String, actual.BackupEmail.String)
	assert.Equal(t, expected.Location.String, actual.Location.String)
}

func TestUserRepository_FindByAccountEmail3(t *testing.T) {
	store := New(configs.NewDatabaseConfig())
	_ = store.Open()
	defer store.Close()
	expected := &models.User{
		UserID:         0,
		AccountEmail:   "blabla@gmail.com",
		PasswordSHA256: "123",
		Username:       "blashka",
		FullName:       "Yurii Ivanitskiy",
		BackupEmail:    &sql.NullString{String: "yi@gmail.com", Valid: true},
		Location:       &sql.NullString{String: "Kiyiv, Ukraine", Valid: true},
	}

	actual, err := store.Users().FindByAccountEmail(expected.AccountEmail)
	if err != nil {
		actual = nil
	}

	assert.Equal(t, expected.AccountEmail, actual.AccountEmail)
	assert.Equal(t, expected.PasswordSHA256, actual.PasswordSHA256)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.FullName, actual.FullName)
	assert.Equal(t, expected.BackupEmail.String, actual.BackupEmail.String)
	assert.Equal(t, expected.Location.String, actual.Location.String)
}

func TestUserRepository_DeleteByID(t *testing.T) {
	store := New(configs.NewDatabaseConfig())
	_ = store.Open()
	defer store.Close()
	expected := &models.User{
		UserID:         0,
		AccountEmail:   "blabla@gmail.com",
		PasswordSHA256: "123",
		Username:       "blashka",
		FullName:       "Yurii Ivanitskiy",
		BackupEmail:    &sql.NullString{String: "yi@gmail.com", Valid: true},
		Location:       &sql.NullString{String: "Kiyiv, Ukraine", Valid: true},
	}

	model, err := store.Users().FindByAccountEmail(expected.AccountEmail)
	assert.Equal(t, nil, err)

	actual, err := store.Users().DeleteByID(model.UserID)
	assert.Equal(t, nil, err)

	assert.Equal(t, expected.AccountEmail, actual.AccountEmail)
	assert.Equal(t, expected.PasswordSHA256, actual.PasswordSHA256)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.FullName, actual.FullName)
	assert.Equal(t, expected.BackupEmail.String, actual.BackupEmail.String)
	assert.Equal(t, expected.Location.String, actual.Location.String)
}
