package sqlxstore

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type RoleRepository struct {
	store *PostgreDatabaseStore
}

func (r *RoleRepository) SelectUserRoles(userID int) ([]models.Role, error) {
	var roles []models.Role
	if err := r.store.db.Select(
		&roles,
		`SELECT * FROM public.roles WHERE role_id IN (SELECT role_id FROM public.user_roles WHERE user_id = $1)`,
		userID,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range roles {
		roles[idx].CheckNullableData()
	}
	return roles, nil
}

func (r *RoleRepository) SelectAll() ([]models.Role, error) {
	var roles []models.Role
	if err := r.store.db.Select(
		&roles,
		`SELECT * FROM public.roles`,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range roles {
		roles[idx].CheckNullableData()
	}
	return roles, nil
}

func (r *RoleRepository) FindByName(roleName string) (*models.Role, error) {
	role := &models.Role{}
	if err := r.store.db.Get(
		role,
		`SELECT * FROM public.roles WHERE name = $1`,
		roleName,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	role.CheckNullableData()
	return role, nil
}

func (r *RoleRepository) FindByID(roleID int) (*models.Role, error) {
	role := &models.Role{}
	if err := r.store.db.Get(
		role,
		`SELECT * FROM public.roles WHERE role_id = $1`,
		roleID,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	role.CheckNullableData()
	return role, nil
}

func (r *RoleRepository) Create(role *models.Role) (*models.Role, error) {
	insertQuery := `
		INSERT INTO public.roles (
                name, description, 
                allow_roles_crud, allow_users_crud, allow_veterinarians_crud, 
                allow_vaccines_crud, allow_food_crud, allow_pets_crud, allow_database_dump)
		VALUES (:name, :description, :allow_roles_crud, :allow_users_crud, :allow_veterinarians_crud, 
                :allow_vaccines_crud, :allow_food_crud, :allow_pets_crud, :allow_database_dump);`

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	role.BeforeCreate()
	if _, err := transaction.NamedExec(insertQuery, role); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	roleModel, err := r.FindByName(role.RoleName)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	roleModel.CheckNullableData()
	return roleModel, nil
}

func (r *RoleRepository) Update(newRole *models.Role) (*models.Role, error) {
	updateQuery := `
		UPDATE public.roles
		SET 
			name = :name,
			description = :description,
			allow_roles_crud = :allow_roles_crud,
			allow_users_crud = :allow_users_crud,
			allow_veterinarians_crud = :allow_veterinarians_crud,
			allow_vaccines_crud = :allow_vaccines_crud,
			allow_food_crud = :allow_food_crud,
			allow_pets_crud = :allow_pets_crud,
			allow_database_dump = :allow_database_dump
		WHERE public.roles.role_id = :role_id;`

	updatingRole, err := r.FindByID(newRole.RoleID)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	updatingRole.Update(newRole)
	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()
	updatingRole.BeforeCreate()
	if _, err := transaction.NamedExec(updateQuery, updatingRole); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return updatingRole, nil
}

func (r *RoleRepository) DeleteByID(roleID int) (*models.Role, error) {
	deletionQuery := `DELETE FROM public.roles WHERE role_id = $1;`
	deletingRole, err := r.FindByID(roleID)
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
	if _, err := transaction.Exec(deletionQuery, roleID); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return deletingRole, nil
}
