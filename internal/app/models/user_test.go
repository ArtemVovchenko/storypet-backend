package models_test

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_BeforeCreate(t *testing.T) {
	testCases := []struct {
		name    string
		u       func() *models.User
		isValid bool
	}{
		{
			name: "Valid",
			u: func() *models.User {
				user := models.TestUser(t)
				return user
			},
			isValid: true,
		},
		{
			name: "Empty Account Email",
			u: func() *models.User {
				user := models.TestUser(t)
				user.AccountEmail = ""
				return user
			},
			isValid: false,
		},
		{
			name: "Valid: Empty Backup Email",
			u: func() *models.User {
				user := models.TestUser(t)
				user.SpecifiedBackupEmail = ""
				return user
			},
			isValid: true,
		},
	}

	for _, tc := range testCases {
		if tc.isValid {
			assert.NoError(t, tc.u().Validate())
		} else {
			assert.Error(t, tc.u().Validate())
		}
	}
}

func TestUser_Validate(t *testing.T) {
	u := models.TestUser(t)
	assert.NoError(t, u.Validate())
}
