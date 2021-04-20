package sessions

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type Session struct {
	UserID      int           `json:"user_id"`
	RefreshUUID string        `json:"refresh_uuid"`
	Roles       []models.Role `json:"roles"`
}
