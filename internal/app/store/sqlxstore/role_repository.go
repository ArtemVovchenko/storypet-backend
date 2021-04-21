package sqlxstore

import "github.com/ArtemVovchenko/storypet-backend/internal/app/models"

type RoleRepository struct {
	store *PostgreDatabaseStore
}

func (r *RoleRepository) SelectUserRoles(userID int) ([]models.Role, error) {
	var roles []models.Role
	if err := r.store.db.Select(
		&roles,
		`SELECT * FROM roles WHERE role_id IN (SELECT role_id FROM user_roles WHERE user_id = $1)`,
		userID,
	); err != nil {
		return nil, err
	}
	for idx := range roles {
		roles[idx].CheckNullableData()
	}
	return roles, nil
}
