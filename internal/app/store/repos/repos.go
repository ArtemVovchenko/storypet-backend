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

	AssignRole(userID int, roleID int) error
	DeleteRole(userID int, roleID int) error

	SelectClinicByUserID(userID int) (*models.VetClinic, error)
	CreateClinic(clinic *models.VetClinic) (*models.VetClinic, error)
	UpdateClinic(clinic *models.VetClinic) (*models.VetClinic, error)
	DeleteClinic(userID int) (*models.VetClinic, error)
}

type RoleRepository interface {
	SelectUserRoles(userID int) ([]models.Role, error)
	SelectAll() ([]models.Role, error)
	FindByName(roleName string) (*models.Role, error)
	FindByID(roleID int) (*models.Role, error)
	Create(role *models.Role) (*models.Role, error)
	Update(newRole *models.Role) (*models.Role, error)
	DeleteByID(roleID int) (*models.Role, error)
}

type PetRepository interface {
	SelectAll() ([]models.Pet, error)
	FindByNameAndOwner(name string, ownerID int) (*models.Pet, error)
	FindByID(petID int) (*models.Pet, error)
	CreatePet(pet *models.Pet) (*models.Pet, error)
	UpdatePet(pet *models.Pet) (*models.Pet, error)
	DeleteByID(petID int) (*models.Pet, error)
	AssignVeterinarian(petID int, veterinarianID int) error
	DeleteVeterinarian(petID int) error
	SpecifyParents(fatherID *int, motherID *int, petID int) error
	RemoveParents(petID int) error
	VerifyMother(petID int) error
	VerifyFather(petID int) error

	SelectAllTypes() ([]models.PetType, error)
	FindTypeByID(typeID int) (*models.PetType, error)
	FindTypeByName(typeName string) (*models.PetType, error)
	CreatePetType(petType *models.PetType) (*models.PetType, error)
	UpdatePetType(other *models.PetType) (*models.PetType, error)
	DeleteTypeByID(typeID int) (*models.PetType, error)

	FindAnthropometryRecordByID(aID int) (*models.Anthropometry, error)
	SelectPetAnthropometryRecords(petID int) ([]models.Anthropometry, error)
	SpecifyAnthropometry(anthropometry *models.Anthropometry) (*models.Anthropometry, error)
	UpdateAnthropometry(anthropometry *models.Anthropometry) (*models.Anthropometry, error)
	DeleteAnthropometryByID(aID int) (*models.Anthropometry, error)
}

type DumpRepository interface {
	Make(savePath string) (*models.Dump, error)
	InsertNewDumpFile(savePath string) (*models.Dump, error)
	Execute(dumpFilePath string) error
	SelectAll() ([]models.Dump, error)
	SelectByName(dumpFileName string) (*models.Dump, error)
	DeleteByName(dumpFileName string) (*models.Dump, error)
}
