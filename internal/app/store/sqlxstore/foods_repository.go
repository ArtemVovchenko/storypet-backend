package sqlxstore

import (
	"database/sql"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"strings"
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
