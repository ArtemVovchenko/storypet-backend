package sqlxstore

import (
	"database/sql"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
	"time"
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
		_ = transaction.Rollback()
	}(transaction)

	_, err = transaction.NamedExec(
		`INSERT INTO public.users 
    			  (account_email, password_sha256, username, full_name, backup_email, location, registration_date)
			   VALUES 
			      (:account_email, :password_sha256, :username, :full_name, :backup_email, :location, :registration_date) 
			   RETURNING user_id`,
		*u,
	)
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
		_ = transaction.Rollback()
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
	userEntity.AfterCreate()
	return userEntity, nil
}

func (r *UserRepository) SelectAll() ([]models.User, error) {
	var userModels []models.User
	if err := r.store.db.Select(&userModels, `SELECT * FROM public.users;`); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range userModels {
		userModels[idx].AfterCreate()
	}
	return userModels, nil
}

func (r *UserRepository) Update(other *models.User) (*models.User, error) {
	updateQuery := `
		UPDATE public.users 
		SET 
			account_email = :account_email,
			username = :username,
			full_name = :full_name,
			backup_email = :backup_email,
			location = :location
		WHERE user_id = :user_id`
	current, err := r.store.Users().FindByID(other.UserID)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	current.Update(other)
	if err := current.BeforeCreate(); err != nil {
		return nil, err
	}
	if err := current.Validate(); err != nil {
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

	if _, err := transaction.NamedExec(
		updateQuery,
		current,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	current.AfterCreate()
	return current, nil
}

func (r *UserRepository) ChangePassword(userID int, newPassword string) error {
	userModel, err := r.FindByID(userID)
	if err != nil {
		return err
	}
	userModel.AfterCreate()
	userModel.Password = newPassword
	if err := userModel.BeforeCreate(); err != nil {
		r.store.logger.Println()
		return err
	}
	if err := userModel.Validate(); err != nil {
		return err
	}

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(
		`UPDATE public.users 
				SET  
					password_sha256 = :password_sha256
				WHERE user_id = :user_id`,
		userModel,
	); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *UserRepository) AssignRole(userID int, roleID int) error {
	var userCurrentRoleID int
	if err := r.store.db.Get(
		userCurrentRoleID,
		`SELECT role_id FROM public.user_roles WHERE user_id = $1`,
		userID,
	); err != nil {
		r.store.logger.Println(err)
		return err
	}
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()
	if _, err := r.store.db.Exec(
		`DELETE FROM public.user_roles WHERE user_id = $1;`,
		userID,
	); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if _, err := r.store.db.Exec(
		`INSERT INTO public.user_roles (user_id, role_id) VALUES ($1, $2);`,
		userID,
		roleID,
	); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if userCurrentRoleID == 3 {
		if _, err := r.store.db.Exec(
			`UPDATE public.users SET subscription_date = $1 WHERE user_id = $2;`,
			time.Now(),
			userID,
		); err != nil {
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

func (r *UserRepository) DeleteRole(userID int, roleID int) error {
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func() {
		_ = transaction.Rollback()
	}()
	if _, err := r.store.db.Exec(
		`DELETE FROM public.user_roles WHERE user_id = $1 AND role_id = $2 ;`,
		userID,
		roleID,
	); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *UserRepository) SelectClinicByUserID(userID int) (*models.VetClinic, error) {
	selectQuery := `SELECT * FROM public.veterinarians_clinic WHERE user_id = $1;`
	model := &models.VetClinic{}
	if err := r.store.db.Get(model, selectQuery, userID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return model, nil
}

func (r *UserRepository) CreateClinic(clinic *models.VetClinic) (*models.VetClinic, error) {
	insertQuery := `
		INSERT INTO 
			public.veterinarians_clinic (user_id, clinic_name, clinic_id) 
		VALUES 
			(:user_id, :clinic_name, :clinic_id);`

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(insertQuery, clinic); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	created, err := r.SelectClinicByUserID(clinic.UserID)
	return created, err
}

func (r *UserRepository) UpdateClinic(clinic *models.VetClinic) (*models.VetClinic, error) {
	updateQuery := `
		UPDATE public.veterinarians_clinic
		SET 
			clinic_id = :clinic_id,
			clinic_name = :clinic_name
		WHERE user_id = :user_id;`
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	if _, err := transaction.NamedExec(updateQuery, clinic); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	updated, err := r.SelectClinicByUserID(clinic.UserID)
	return updated, err
}

func (r *UserRepository) DeleteClinic(userID int) (*models.VetClinic, error) {
	deleteQuery := `DELETE FROM public.veterinarians_clinic WHERE user_id = $1`

	deletedModel, err := r.SelectClinicByUserID(userID)
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

	if _, err := transaction.Exec(deleteQuery, userID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	return deletedModel, nil
}
