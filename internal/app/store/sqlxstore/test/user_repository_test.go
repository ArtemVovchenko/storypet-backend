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

func TestChangePassword(t *testing.T) {
	userModel, _ := store.Users().FindByAccountEmail("sebre.ds@gmail.com")
	testCases := []struct {
		name     string
		id       int
		password string
		isError  bool
	}{
		{
			name:     "Not Valid Password (Empty)",
			id:       userModel.UserID,
			password: "",
			isError:  false,
		},
		{
			name:     "valid",
			id:       userModel.UserID,
			password: "qwertyqwertyqwerty",
			isError:  false,
		},
		{
			name:     "Not Valid Password (Short)",
			id:       userModel.UserID,
			password: "1",
			isError:  true,
		},
	}
	for _, tc := range testCases {
		err := store.Users().ChangePassword(tc.id, tc.password)
		if tc.isError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
	newModel, _ := store.Users().FindByID(userModel.UserID)
	assert.NotEqualf(t, userModel.PasswordSHA256, newModel.PasswordSHA256, "Password still equals")
}

func TestUpdate(t *testing.T) {
	validUser, err := store.Users().FindByAccountEmail("sebre.ds@gmail.com")
	assert.NoError(t, err)
	testCases := []struct {
		name    string
		entity  models.User
		isError bool
	}{
		{
			name: "No account email",
			entity: models.User{
				UserID:            validUser.UserID,
				AccountEmail:      "",
				Password:          "qwerty123",
				Username:          "sebreID",
				FullName:          "Sebre Adjando",
				SpecifiedLocation: "Colorado, USA",
			},
			isError: false,
		},
		{
			name: "Valid",
			entity: models.User{
				UserID:            validUser.UserID,
				AccountEmail:      "sebre.ds@gmail.us",
				Password:          "qwerty",
				Username:          "sebreID33",
				FullName:          "Sebre Adjando",
				SpecifiedLocation: "New York, USA",
			},
			isError: false,
		},
		{
			name: "Duplicated email",
			entity: models.User{
				UserID:            validUser.UserID,
				AccountEmail:      "vovchenko.artem@icloud.com",
				Password:          "qwerty123",
				Username:          "sebreID",
				FullName:          "Sebre Adjando",
				SpecifiedLocation: "Colorado, USA",
			},
			isError: true,
		},
		{
			name: "Duplicated username",
			entity: models.User{
				UserID:            validUser.UserID,
				AccountEmail:      "sebre.ds@gmail.us",
				Password:          "qwerty123",
				Username:          "an_unseen_future",
				FullName:          "Sebre Adjando",
				SpecifiedLocation: "Colorado, USA",
			},
			isError: true,
		},
	}
	for _, tc := range testCases {
		_, err := store.Users().Update(&tc.entity)
		if tc.isError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestDeleteByID(t *testing.T) {
	validUser, err := store.Users().FindByAccountEmail("sebre.ds@gmail.us")
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
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
