package sqlxstore

import (
	"database/sql"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"strings"
	"time"
)

type FoodRepository struct {
	store *PostgreDatabaseStore
}

func (r *FoodRepository) SelectAll() ([]models.Food, error) {
	query := `SELECT * FROM public.food;`
	var foodModels []models.Food
	if err := r.store.db.Select(&foodModels, query); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range foodModels {
		foodModels[idx].AfterCreate()
	}
	return foodModels, nil
}

func (r *FoodRepository) FindByID(foodID int) (*models.Food, error) {
	query := `SELECT * FROM public.food WHERE food_id = $1;`
	foodModel := &models.Food{}
	if err := r.store.db.Get(foodModel, query, foodID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	foodModel.AfterCreate()
	return foodModel, nil
}

func (r *FoodRepository) SelectByNameSimilarity(namePattern string) ([]models.Food, error) {
	query := `SELECT * FROM public.food WHERE LOWER(food_name) LIKE $1;`
	var foodModels []models.Food
	pattern := "%" + strings.ToLower(namePattern) + "%"

	if err := r.store.db.Select(&foodModels, query, pattern); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return foodModels, nil
		}
		r.store.logger.Println(err)
		return nil, err
	}

	for idx := range foodModels {
		foodModels[idx].AfterCreate()
	}
	return foodModels, nil
}

func (r *FoodRepository) Create(foodModel *models.Food) (*models.Food, error) {
	query := `
		INSERT INTO public.food 
			(food_name, calories, description, manufacturer, creator_id) 
		VALUES 
			($1, $2, $3, $4, $5) 
		RETURNING food_id;`
	var newModelID int
	foodModel.BeforeCreate()

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if err := transaction.QueryRowx(query,
		foodModel.FoodName,
		foodModel.Calories,
		foodModel.Description,
		foodModel.Manufacturer,
		foodModel.CreatorID,
	).Scan(&newModelID); err != nil {
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	foodModel.FoodID = newModelID
	return foodModel, nil
}

func (r *FoodRepository) Update(foodModel *models.Food) (*models.Food, error) {
	query := `
		UPDATE public.food 
		SET 
			food_name = :food_name, 
			calories = :calories, 
			description = :description, 
			manufacturer = :manufacturer,
		    creator_id = :creator_id
		WHERE food_id = :food_id;`
	foodModel.BeforeCreate()
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(query, foodModel); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	return r.FindByID(foodModel.FoodID)
}

func (r *FoodRepository) DeleteByID(foodID int) (*models.Food, error) {
	query := `DELETE FROM public.food WHERE food_id = $1;`
	foodModel, err := r.FindByID(foodID)
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

	if _, err := transaction.Exec(query, foodID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return foodModel, nil
}

func (r *FoodRepository) AddPetEating(eating *models.Eating) error {
	query := `INSERT INTO public.eatings (eating_timestamp, pet_id, food_id) VALUES (:eating_timestamp, :pet_id, :food_id);`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(query, eating); err != nil {
		r.store.logger.Println(err)
		return err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *FoodRepository) GetPetsEatingsForDate(petID int, date time.Time) ([]models.Food, error) {
	query := `
		SELECT f.food_id, food_name, calories, description, manufacturer FROM public.eatings 
		JOIN public.food f 
		ON f.food_id = eatings.food_id
		WHERE pet_id = $1 AND eating_timestamp::date = $2;`

	var foods []models.Food
	if err := r.store.db.Select(&foods, query, petID, date); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range foods {
		foods[idx].AfterCreate()
	}
	return foods, nil
}
