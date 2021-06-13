package sqlxstore

import (
	"database/sql"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"time"
)

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

func (r *PetRepository) SelectByUserID(userID int) ([]models.Pet, error) {
	var petModels []models.Pet
	if err := r.store.db.Select(&petModels, `SELECT * FROM public.pets WHERE user_id = $1;`, userID); err != nil {
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
		WHERE public.pets.pet_id = :pet_id`

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

func (r *PetRepository) SpecifyParents(fatherID *int, motherID *int, petID int) error {
	spMother := `UPDATE public.pets SET mother_id = $1, mother_verified = FALSE WHERE pet_id = $2;`
	spFather := `UPDATE public.pets SET father_id = $1, father_verified = FALSE WHERE pet_id = $2;`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if motherID != nil {
		if _, err := transaction.Exec(spMother, *motherID, petID); err != nil {
			r.store.logger.Println(err)
			return err
		}
	}

	if fatherID != nil {
		if _, err := transaction.Exec(spFather, *fatherID, petID); err != nil {
			r.store.logger.Println(err)
			return err
		}
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *PetRepository) RemoveParents(petID int) error {
	query := `
		UPDATE public.pets 
		SET 
			father_id = NULL, 
			father_verified = FALSE,
			mother_id = NULL,
			mother_verified = FALSE
		WHERE pet_id = $1;`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.Exec(query, petID); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *PetRepository) VerifyMother(petID int) error {
	query := `UPDATE public.pets SET mother_verified = TRUE WHERE pet_id = $1;`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.Exec(query, petID); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *PetRepository) VerifyFather(petID int) error {
	query := `UPDATE public.pets SET father_verified = TRUE WHERE pet_id = $1;`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.Exec(query, petID); err != nil {
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

func (r *PetRepository) FindTypeByName(typeName string) (*models.PetType, error) {
	selectQuery := `SELECT * FROM public.pet_types WHERE type_name = $1;`
	petType := &models.PetType{}
	if err := r.store.db.Get(petType, selectQuery, typeName); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petType, nil
}

func (r *PetRepository) FindTypeByID(typeID int) (*models.PetType, error) {
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

	return r.FindTypeByName(petType.TypeName)
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

	return r.FindTypeByID(other.TypeID)
}

func (r *PetRepository) DeleteTypeByID(typeID int) (*models.PetType, error) {
	deleteQuery := `DELETE FROM public.pet_types WHERE type_id = $1;`
	deletingType, err := r.FindTypeByID(typeID)
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

func (r *PetRepository) FindAnthropometryRecordByID(aID int) (*models.Anthropometry, error) {
	query := `SELECT * FROM public.anthropometries WHERE record_id = $1`
	aModel := &models.Anthropometry{}
	if err := r.store.db.Get(aModel, query, aID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return aModel, nil
}

func (r *PetRepository) SelectPetAnthropometryRecords(petID int) ([]models.Anthropometry, error) {
	query := `SELECT * FROM public.anthropometries WHERE pet_id = $1 ORDER BY record_time DESC;`
	var aModels []models.Anthropometry
	if err := r.store.db.Select(&aModels, query, petID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return aModels, nil
}

func (r *PetRepository) SpecifyAnthropometry(anthropometry *models.Anthropometry) (*models.Anthropometry, error) {
	query := `INSERT INTO public.anthropometries (pet_id, record_time, height, weight) VALUES ($1, $2, $3, $4) RETURNING record_id;`
	var anthropometryID int

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if err := transaction.QueryRowx(
		query,
		anthropometry.PetID,
		anthropometry.Time,
		anthropometry.Height,
		anthropometry.Weight,
	).Scan(&anthropometryID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return r.FindAnthropometryRecordByID(anthropometryID)
}

func (r *PetRepository) UpdateAnthropometry(anthropometry *models.Anthropometry) (*models.Anthropometry, error) {
	query := `UPDATE public.anthropometries SET height = :height, weight = :weight WHERE record_id = :record_id;`

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(query, anthropometry); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return r.FindAnthropometryRecordByID(anthropometry.RecordID)
}

func (r *PetRepository) DeleteAnthropometryByID(aID int) (*models.Anthropometry, error) {
	query := `DELETE FROM public.anthropometries WHERE record_id = $1;`
	deletingModel, err := r.FindAnthropometryRecordByID(aID)
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

	if _, err := transaction.Exec(query, aID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	return deletingModel, nil
}

func (r *PetRepository) CreateActivityRecord(record *models.Activity) error {
	query := `
		INSERT INTO public.activity
			(record_timestamp, pet_id, distance, mean_speed)
		VALUES
			(:record_timestamp, :pet_id, :distance, :mean_speed);`

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(query, record); err != nil {
		r.store.logger.Println(err)
		return err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *PetRepository) SelectPetActivityRecords(petID int) ([]models.Activity, error) {
	query := `SELECT * FROM public.activity WHERE pet_id = $1;`
	var petActivityModels []models.Activity
	if err := r.store.db.Select(&petActivityModels, query, petID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petActivityModels, nil
}

func (r *PetRepository) SelectPetActivityRecordsInInterval(petID int, start time.Time, end time.Time) ([]models.Activity, error) {
	query := `SELECT * FROM public.activity WHERE pet_id = $1 AND record_timestamp::date >= $2 AND record_timestamp::date <= $3;`
	var petActivityModels []models.Activity
	if err := r.store.db.Select(&petActivityModels, query, petID, start, end); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petActivityModels, nil
}

func (r *PetRepository) SelectPetActivityRecordsToTime(petID int, start time.Time) ([]models.Activity, error) {
	query := `SELECT * FROM public.activity WHERE pet_id = $1 AND record_timestamp::date <= $2;`
	var petActivityModels []models.Activity
	if err := r.store.db.Select(&petActivityModels, query, petID, start); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return petActivityModels, nil
}

func (r *PetRepository) GetPetStatistics(petID int) (
	[]models.FoodCaloriesReport,
	[]models.RERCaloriesReport,
	[]models.AnthropometryReport,
	[]models.ActivityReport,
	error) {
	foodQuery := `
		SELECT date(eating_timestamp), SUM(calories * eatings.portion_weight) AS "eat_ccal" FROM eatings
		INNER JOIN food f ON f.food_id = eatings.food_id
		WHERE pet_id = $1
		GROUP BY date(eating_timestamp);`
	RERQuery := `
		SELECT date(a.record_time), 70 * pt.rer_coefficient * POWER(a.weight, 0.75) as rer_ccal FROM pets
		INNER JOIN anthropometries a on pets.pet_id = a.pet_id
		INNER JOIN pet_types pt on pt.type_id = pets.pet_type
		WHERE a.pet_id = $1;`
	anthropometryQuery := `
		SELECT DATE(record_time) as date, height, weight FROM anthropometries WHERE pet_id = $1;`
	activityQuery := `
		SELECT date(record_timestamp) as date, SUM(distance) as distance, AVG(mean_speed) as mean_speed FROM public.activity
		WHERE pet_id = $1
		GROUP BY date(record_timestamp);`

	var foodModels []models.FoodCaloriesReport
	var rerModels []models.RERCaloriesReport
	var anthropometryModels []models.AnthropometryReport
	var activityModels []models.ActivityReport

	if err := r.store.db.Select(&foodModels, foodQuery, petID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.store.logger.Println(err)
		return nil, nil, nil, nil, err
	}

	if err := r.store.db.Select(&rerModels, RERQuery, petID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.store.logger.Println(err)
		return nil, nil, nil, nil, err
	}

	if err := r.store.db.Select(&anthropometryModels, anthropometryQuery, petID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.store.logger.Println(err)
		return nil, nil, nil, nil, err
	}

	if err := r.store.db.Select(&activityModels, activityQuery, petID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.store.logger.Println(err)
		return nil, nil, nil, nil, err
	}
	return foodModels, rerModels, anthropometryModels, activityModels, nil
}

func (r *PetRepository) GetPetDateStatistics(petID int, day time.Time) (*models.TodayReport, error) {
	type activityResult struct {
		Distance  float64 `db:"distance"`
		MeanSpeed float64 `db:"mean_speed"`
	}

	currentFoodCaloriesQuery := `
		SELECT SUM(calories * eatings.portion_weight) AS "eat_ccal" FROM eatings
		INNER JOIN food f ON f.food_id = eatings.food_id
		WHERE pet_id = $1 AND eating_timestamp::date = $2
		GROUP BY date(eating_timestamp);`
	currentRERCaloriesQuery := `
		SELECT 70 * (
			SELECT rer_coefficient FROM pet_types INNER JOIN pets p on pet_types.type_id = p.pet_type WHERE pet_id = $1
			) *
			power(
				(SELECT weight FROM anthropometries WHERE pet_id = $1 ORDER BY record_time DESC LIMIT 1),
				0.75
			) AS "rer_ccal";`
	currentActivityQuery := `
		SELECT SUM(distance) as distance, AVG(mean_speed) as mean_speed FROM public.activity 
		WHERE pet_id = $1 AND record_timestamp::date = $2
		GROUP BY date(record_timestamp);`

	var currentFoodCal float64
	var currentRERCal float64
	currentActivity := &activityResult{}

	if err := r.store.db.Get(&currentFoodCal, currentFoodCaloriesQuery, petID, day); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	if err := r.store.db.Get(&currentRERCal, currentRERCaloriesQuery, petID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	if err := r.store.db.Get(currentActivity, currentActivityQuery, petID, day); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	return &models.TodayReport{
		FoodTotalCalories: currentFoodCal,
		RERTotalCalories:  currentRERCal,
		MeanSpeed:         currentActivity.MeanSpeed,
		TotalDistance:     currentActivity.Distance,
	}, nil
}

func (r *PetRepository) CreatePetHealthReport(report *models.PetHealthReport) error {
	insertQuery := `
		INSERT INTO public.pet_health_reports (pet_id, veterinarian_id, report_timestamp, report_conclusion, report_comments)
		VALUES (:pet_id, :veterinarian_id, :report_timestamp, :report_conclusion, :report_comments);`

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(insertQuery, report); err != nil {
		r.store.logger.Println(err)
		return err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *PetRepository) GetAllPetHealthReports(petID int) ([]models.PetHealthReport, error) {
	query := `SELECT * FROM public.pet_health_reports WHERE pet_id = $1;`
	var reports []models.PetHealthReport
	if err := r.store.db.Select(&reports, query, petID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range reports {
		reports[idx].AfterCreate()
	}
	return reports, nil
}
