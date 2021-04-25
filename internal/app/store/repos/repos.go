package repos

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type UserRepository interface {
	Create(u *models.User) (*models.User, error)
	DeleteByID(id int) (*models.User, error)
	FindByAccountEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error)
	SelectAll() ([]models.User, error)
	Update(other *models.User) (*models.User, error)
	ChangePassword(userID int, newPassword string) error
}

type RoleRepository interface {
	SelectUserRoles(userID int) ([]models.Role, error)
}

type DumpRepository interface {
	Make(savePath string) (*models.Dump, error)
	InsertNewDumpFile(savePath string) (*models.Dump, error)
	Execute(dumpFilePath string) error
	SelectAll() ([]models.Dump, error)
	SelectByName(dumpFileName string) (*models.Dump, error)
	DeleteByName(dumpFileName string) (*models.Dump, error)
}
