package sqlxstore

import (
	"database/sql"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
)

type IoTDevicesRepository struct {
	store *PostgreDatabaseStore
}

func (r *IoTDevicesRepository) GetByID(deviceID int) (*models.IoTDevice, error) {
	query := `SELECT * FROM public.iot_devices WHERE device_id = $1;`
	deviceModel := &models.IoTDevice{}
	if err := r.store.db.Get(deviceModel, query, deviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		r.store.logger.Println("Database Error: %v: ", err)
		return nil, err
	}
	return deviceModel, nil
}

func (r *IoTDevicesRepository) GetByAccessSecret(accessSecret string) (*models.IoTDevice, error) {
	query := `SELECT * FROM public.iot_devices WHERE access_secret = $1;`
	deviceModel := &models.IoTDevice{}
	if err := r.store.db.Get(deviceModel, query, accessSecret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		r.store.logger.Println("Database Error: %v: ", err)
		return nil, err
	}
	return deviceModel, nil
}
