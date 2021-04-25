package sqlxstore

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreate(t *testing.T) {
	testCases := []struct {
		name    string
		entity  models.User
		isError bool
	}{
		{
			name: "Invalid no password",
			entity: models.User{
				AccountEmail:      "sebre.ds@gmail.com",
				Username:          "sebreID",
				FullName:          "Sebre Adjando",
				SpecifiedLocation: "Colorado, USA",
			},
			isError: true,
		},
		{
			name: "No account email",
			entity: models.User{
				Password:          "qwerty123",
				Username:          "sebreID",
				FullName:          "Sebre Adjando",
				SpecifiedLocation: "Colorado, USA",
			},
			isError: true,
		},
		{
			name: "Valid",
			entity: models.User{
				AccountEmail:      "sebre.ds@gmail.com",
				Password:          "qwerty123",
				Username:          "sebreID",
				FullName:          "Sebre Adjando",
				SpecifiedLocation: "Colorado, USA",
			},
			isError: false,
		},
		{
			name: "duplicated",
			entity: models.User{
				AccountEmail:      "sebre.ds@gmail.com",
				Password:          "qwerty123",
				Username:          "sebreID",
				FullName:          "Sebre Adjando",
				SpecifiedLocation: "Colorado, USA",
			},
			isError: true,
		},
	}

	for _, tc := range testCases {
		_, err := store.Users().Create(&tc.entity)
		if tc.isError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestFindByAccountEmail(t *testing.T) {
	testCases := []struct {
		name         string
		accountEmail string
		isError      bool
	}{
		{
			name:         "No Such account",
			accountEmail: "vovchenko.artem@iclod.com",
			isError:      true,
		},
		{
			name:         "No such account 2",
			accountEmail: "sebr.ds@gmail.com",
			isError:      true,
		},
		{
			name:         "Valid",
			accountEmail: "sebre.ds@gmail.com",
			isError:      false,
		},
	}

	for _, tc := range testCases {
		model, err := store.Users().FindByAccountEmail(tc.accountEmail)
		if tc.isError {
			assert.Error(t, err)
			assert.Nil(t, model)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, model)
		}
	}
}

func TestFindByID(t *testing.T) {
	validUser, err := store.Users().FindByAccountEmail("sebre.ds@gmail.com")
	assert.NoError(t, err)
	testCases := []struct {
		name    string
		id      int
		isError bool
	}{
		{
			name:    "No Such account",
			id:      14,
			isError: true,
		},
		{
			name:    "No such account 2",
			id:      32,
			isError: true,
		},
		{
			name:    "Valid",
			id:      validUser.UserID,
			isError: false,
		},
	}

	for _, tc := range testCases {
		model, err := store.Users().FindByID(tc.id)
		if tc.isError {
			assert.Error(t, err)
			assert.Nil(t, model)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, model)
		}
	}
}

func TestDeleteByID(t *testing.T) {
	validUser, err := store.Users().FindByAccountEmail("sebre.ds@gmail.com")
	assert.NoError(t, err)
	testCases := []struct {
		name    string
		id      int
		isError bool
	}{
		{
			name:    "No Such account",
			id:      14,
			isError: true,
		},
		{
			name:    "No such account 2",
			id:      32,
			isError: true,
		},
		{
			name:    "Valid",
			id:      validUser.UserID,
			isError: false,
		},
	}

	for _, tc := range testCases {
		model, err := store.Users().DeleteByID(tc.id)
		if tc.isError {
			assert.Error(t, err)
			assert.Nil(t, model)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, model)
		}
	}

	_, err = store.Users().FindByID(validUser.UserID)
	assert.Error(t, err)
}

func TestSelectAll(t *testing.T) {
	_, err := store.Users().SelectAll()
	assert.NoError(t, err)
}
