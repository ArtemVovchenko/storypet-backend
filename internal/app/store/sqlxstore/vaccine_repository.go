package sqlxstore

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type VaccineRepository struct {
	store *PostgreDatabaseStore
}

func (r *VaccineRepository) FindByID(vaccineID int) (*models.Vaccine, error) {
	query := `SELECT * FROM public.vaccines WHERE vaccine_id = $1;`
	vaccine := &models.Vaccine{}
	if err := r.store.db.Get(vaccine, query, vaccineID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	vaccine.AfterCreate()
	return vaccine, nil
}

func (r *VaccineRepository) SelectByPetID(petID int) ([]models.Vaccine, error) {
	query := `SELECT * FROM public.vaccines WHERE pet_id = $1;`
	var vaccines []models.Vaccine
	if err := r.store.db.Select(&vaccines, query, petID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range vaccines {
		vaccines[idx].AfterCreate()
	}
	return vaccines, nil
}

func (r *VaccineRepository) Create(vaccine *models.Vaccine) (*models.Vaccine, error) {
	insertQuery := `
		INSERT INTO public.vaccines 
			(pet_id, name, vaccination_date, description) 
		VALUES
			($1, $2, $3, $4)
		RETURNING vaccine_id;`
	selectQuery := `
		SELECT * FROM public.vaccines
		WHERE vaccine_id = $1;`

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()
	var newVaccineID int
	if err := transaction.QueryRowx(
		insertQuery,
		vaccine.PetID,
		vaccine.Name,
		vaccine.VaccinationDate,
		vaccine.Description,
	).Scan(&newVaccineID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	createdModel := &models.Vaccine{}
	if err := r.store.db.Get(createdModel, selectQuery, newVaccineID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	createdModel.AfterCreate()
	return createdModel, nil
}

func (r *VaccineRepository) Update(vaccine *models.Vaccine) (*models.Vaccine, error) {
	updateQuery := `
		UPDATE public.vaccines 
		SET 
			pet_id = :pet_id, 
			name = :name, 
			vaccination_date = :vaccination_date, 
			description = :description 
		WHERE vaccine_id = :vaccine_id;`
	updatingModel, err := r.FindByID(vaccine.VaccineID)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	updatingModel.Update(vaccine)
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(updateQuery, updatingModel); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return updatingModel, nil
}

func (r *VaccineRepository) DeleteByID(vaccineID int) (*models.Vaccine, error) {
	deletingQuery := `DELETE FROM public.vaccines WHERE vaccine_id = $1;`
	deletingModel, err := r.FindByID(vaccineID)
	if err != nil {
		r.store.logger.Println(err)
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

	if _, err := transaction.Exec(deletingQuery, vaccineID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	deletingModel.AfterCreate()
	return deletingModel, nil
}
