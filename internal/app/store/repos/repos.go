package repos

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type UserRepository interface {
	Create(u *models.User) (*models.User, error)
	DeleteByID(id int) (*models.User, error)
	FindByAccountEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error)
}

type RoleRepository interface {
	SelectUserRoles(userID int) ([]models.Role, error)
}

type DumpRepository interface {
	MakeDump(savePath string) error
	ExecuteDump(dumpQueries string) error
}
