package store

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *models.User) (*models.User, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}

	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}

	transaction, err := r.store.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()

	_, err = transaction.NamedExec(`INSERT INTO users (account_email, password_sha256, username, full_name, backup_email, location)
		VALUES (:account_email, :password_sha256, :username, :full_name, :backup_email, :location) RETURNING user_id`, *u)
	if err != nil {
		return nil, err
	}

	err = transaction.Get(u, `SELECT user_id FROM users WHERE account_email = $1`, u.AccountEmail)
	if err != nil {
		return nil, err
	}

	_, err = transaction.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, (SELECT DISTINCT role_id FROM roles WHERE name = 'unsubscribed'))`,
		u.UserID,
	)
	if err != nil {
		return nil, err
	}

	err = transaction.Commit()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) DeleteByID(id int) (*models.User, error) {
	userModel := &models.User{}

	if err := r.store.db.Get(userModel, `SELECT * FROM users WHERE  user_id = $1`, id); err != nil {
		return nil, err
	}

	transaction, err := r.store.db.Begin()
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()

	_, err = transaction.Exec(`DELETE FROM users WHERE user_id = $1`, userModel.UserID)
	if err != nil {
		return nil, err
	}

	err = transaction.Commit()
	if err != nil {
		return nil, err
	}

	return userModel, nil
}

func (r *UserRepository) FindByAccountEmail(email string) (*models.User, error) {
	userEntity := &models.User{}
	if err := r.store.db.Get(userEntity,
		`SELECT * FROM users WHERE account_email = $1 LIMIT 1`,
		email,
	); err != nil {
		return nil, err
	}

	return userEntity, nil
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	userEntity := &models.User{}
	if err := r.store.db.Get(userEntity,
		`SELECT * FROM users WHERE user_id = $1`,
		id,
	); err != nil {
		return nil, err
	}
	return userEntity, nil
}
