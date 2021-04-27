package sqlxstore

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type PetRepository struct {
	store *PostgreDatabaseStore
}

func (r *PetRepository) FindByNameAndOwner(name string, ownerID int) (*models.Pet, error) {
	selectQuery := `SELECT * FROM public.pets WHERE name = $1 AND user_id = $2;`
	petModel := &models.Pet{}
	if err := r.store.db.Get(petModel, selectQuery, name, ownerID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petModel, nil
}

func (r *PetRepository) FindByID(petID int) (*models.Pet, error) {
	selectQuery := `SELECT * FROM public.pets WHERE pet_id = $1;`
	petModel := &models.Pet{}
	if err := r.store.db.Get(petModel, selectQuery, petID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petModel, nil
}

func (r *PetRepository) CreatePet(pet *models.Pet) (*models.Pet, error) {
	createQuery := `
		INSERT INTO 
			public.pets (name, user_id, veterinar_id, pet_type, breed, family_name) 
		VALUES 
			(:name, :user_id, :veterinar_id, :pet_type, :breed, :family_name)`
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
			user_id = :user_id,
			veterinar_id = :veterinar_id,
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
	return deletingPet, nil
}
