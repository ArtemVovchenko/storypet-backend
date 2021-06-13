package models

type IoTDevice struct {
	DeviceID     int    `json:"device_id" db:"device_id"`
	PetID        int    `json:"pet_id" db:"pet_id"`
	AccessSecret string `json:"access_secret" db:"access_secret"`
}

func (i *IoTDevice) Update(device IoTDevice) {
	i.PetID = device.PetID
	i.AccessSecret = device.AccessSecret
}
