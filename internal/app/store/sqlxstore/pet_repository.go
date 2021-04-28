package sqlxstore

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type PetRepository struct {
	store *PostgreDatabaseStore
}

func (r *PetRepository) SelectAll() ([]models.Pet, error) {
	var petModels []models.Pet
	if err := r.store.db.Select(&petModels, `SELECT * FROM public.pets;`); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range petModels {
		petModels[idx].AfterCreate()
	}
	return petModels, nil
}

func (r *PetRepository) FindByNameAndOwner(name string, ownerID int) (*models.Pet, error) {
	selectQuery := `SELECT * FROM public.pets WHERE name = $1 AND user_id = $2;`
	petModel := &models.Pet{}
	if err := r.store.db.Get(petModel, selectQuery, name, ownerID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	petModel.AfterCreate()
	return petModel, nil
}

func (r *PetRepository) FindByID(petID int) (*models.Pet, error) {
	selectQuery := `SELECT * FROM public.pets WHERE pet_id = $1;`
	petModel := &models.Pet{}
	if err := r.store.db.Get(petModel, selectQuery, petID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	petModel.AfterCreate()
	return petModel, nil
}

func (r *PetRepository) CreatePet(pet *models.Pet) (*models.Pet, error) {
	createQuery := `
		INSERT INTO 
			public.pets (name, user_id, veterinarian_id, pet_type, breed, family_name, mother_id, father_id) 
		VALUES 
			(:name, :user_id, :veterinarian_id, :pet_type, :breed, :family_name, :mother_id, :father_id);`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := r.store.db.NamedExec(createQuery, pet); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return r.FindByNameAndOwner(pet.Name, pet.UserID)
}

func (r *PetRepository) UpdatePet(pet *models.Pet) (*models.Pet, error) {
	updateQuery := `
		UPDATE public.pets
		SET 
			name = :name,
		    pet_type = :pet_type,
			user_id = :user_id,
			veterinarian_id = :veterinarian_id,
			breed = :breed,
			family_name = :family_name,
			mother_id = :mother_id,
			mother_verified = :mother_verified,
			father_id = :father_id,
			father_verified = :father_verified
		WHERE public.pets.pet_id = :pet_id
`

	updatingPet, err := r.FindByID(pet.PetID)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	updatingPet.AfterCreate()
	updatingPet.Update(pet)
	updatingPet.BeforeCreate()

	transaction, err := r.store.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(updateQuery, updatingPet); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}
	updatingPet.AfterCreate()
	return updatingPet, nil
}

func (r *PetRepository) DeleteByID(petID int) (*models.Pet, error) {
	deleteQuery := `DELETE FROM public.pets WHERE pet_id = $1;`
	deletingPet, err := r.FindByID(petID)
	if err != nil {
		return nil, err
	}
	transaction, err := r.store.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.Exec(deleteQuery, petID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}
	deletingPet.AfterCreate()
	return deletingPet, nil
}

func (r *PetRepository) AssignVeterinarian(petID int, veterinarianID int) error {
	query := `UPDATE public.pets SET veterinarian_id = $1 WHERE pet_id = $2`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := r.store.db.Exec(query, veterinarianID, petID); err != nil {
		r.store.logger.Println(err)
		return err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *PetRepository) DeleteVeterinarian(petID int) error {
	query := `UPDATE public.pets SET veterinarian_id = NULL WHERE pet_id = $1`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := r.store.db.Exec(query, petID); err != nil {
		r.store.logger.Println(err)
		return err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *PetRepository) SelectAllTypes() ([]models.PetType, error) {
	selectQuery := `SELECT * FROM public.pet_types;`
	var petTypes []models.PetType
	if err := r.store.db.Select(&petTypes, selectQuery); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petTypes, nil
}

func (r *PetRepository) SelectTypeByName(typeName string) (*models.PetType, error) {
	selectQuery := `SELECT * FROM public.pet_types WHERE type_name = $1;`
	petType := &models.PetType{}
	if err := r.store.db.Get(petType, selectQuery, typeName); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petType, nil
}

func (r *PetRepository) SelectTypeByID(typeID int) (*models.PetType, error) {
	selectQuery := `SELECT * FROM public.pet_types WHERE type_id = $1;`
	petType := &models.PetType{}
	if err := r.store.db.Get(petType, selectQuery, typeID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petType, nil
}

func (r *PetRepository) CreatePetType(petType *models.PetType) (*models.PetType, error) {
	insertQuery := `
		INSERT INTO public.pet_types 
			(type_name, rer_coefficient) 
		VALUES 
			(:type_name, :rer_coefficient)`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(insertQuery, petType); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return r.SelectTypeByName(petType.TypeName)
}

func (r *PetRepository) UpdatePetType(other *models.PetType) (*models.PetType, error) {
	updateQuery := `
		UPDATE public.pet_types
		SET 
			type_name = :type_name,
			rer_coefficient = :rer_coefficient
		WHERE type_id = :type_id`

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(updateQuery, other); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	return r.SelectTypeByID(other.TypeID)
}

func (r *PetRepository) DeleteTypeByID(typeID int) (*models.PetType, error) {
	deleteQuery := `DELETE FROM public.pet_types WHERE type_id = $1;`
	deletingType, err := r.SelectTypeByID(typeID)
	if err != nil {
		return nil, err
	}

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.Exec(deleteQuery, typeID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return deletingType, nil
}
