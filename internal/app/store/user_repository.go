package store

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *models.User) (*models.User, error) {
	return nil, nil
}

func (r *UserRepository) FindByAccountEmail(email string) (*models.User, error) {
	userEntity := &models.User{}
	if err := r.store.db.Get(userEntity, "SELECT * FROM users WHERE account_email = $1 LIMIT 1", email); err != nil {
		return nil, err
	}

	return userEntity, nil
}
