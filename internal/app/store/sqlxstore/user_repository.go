package sqlxstore

import (
	"database/sql"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	store *PostgreDatabaseStore
}

func (r *UserRepository) Create(u *models.User) (*models.User, error) {
	if err := u.Validate(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := u.BeforeCreate(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func(transaction *sqlx.Tx) {
		if err := transaction.Rollback(); err != nil {
			r.store.logger.Println(err)
		}
	}(transaction)

	_, err = transaction.NamedExec(`INSERT INTO public.users (account_email, password_sha256, username, full_name, backup_email, location)
		VALUES (:account_email, :password_sha256, :username, :full_name, :backup_email, :location) RETURNING user_id`, *u)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	err = transaction.Get(u, `SELECT user_id FROM public.users WHERE account_email = $1`, u.AccountEmail)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	_, err = transaction.Exec(`INSERT INTO public.user_roles (user_id, role_id) VALUES ($1, (SELECT DISTINCT default_user_role_id FROM public.config LIMIT 1))`,
		u.UserID,
	)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	err = transaction.Commit()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) DeleteByID(id int) (*models.User, error) {
	userModel := &models.User{}

	if err := r.store.db.Get(userModel, `SELECT * FROM users WHERE  user_id = $1`, id); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	transaction, err := r.store.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func(transaction *sql.Tx) {
		if err := transaction.Rollback(); err != nil {
			r.store.logger.Println(err)
		}

	}(transaction)

	_, err = transaction.Exec(`DELETE FROM public.users WHERE user_id = $1`, userModel.UserID)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	err = transaction.Commit()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	return userModel, nil
}

func (r *UserRepository) FindByAccountEmail(email string) (*models.User, error) {
	userEntity := &models.User{}
	if err := r.store.db.Get(userEntity,
		`SELECT * FROM public.users WHERE account_email = $1 LIMIT 1`,
		email,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	return userEntity, nil
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	userEntity := &models.User{}
	if err := r.store.db.Get(userEntity,
		`SELECT * FROM public.users WHERE user_id = $1`,
		id,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return userEntity, nil
}
