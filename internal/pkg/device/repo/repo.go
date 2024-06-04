package repo

import (
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
)

// Repository type definition
type Repository struct {
	// accClient accountsvc.AccountServiceClient
	source dbdevice.DBSource
	mapper
}

// GetByID returns a device of given id.
func (r *Repository) GetByID(deviceID int64) (*device.Device, error) {
	dbDevice, err := r.source.GetByID(deviceID)
	if dbDevice != nil {
		return r.mapper.dbDeviceToDevice(*dbDevice), nil
	}
	return nil, err
}

// GetByAppIDAndIFA returns a device of given appID and ifa.
func (r *Repository) GetByAppIDAndIFA(appID int64, ifa string) (*device.Device, error) {
	dbDevice, err := r.source.GetByAppIDAndIFA(appID, ifa)
	if dbDevice != nil {
		return r.mapper.dbDeviceToDevice(*dbDevice), nil
	}
	return nil, err
}

// GetByAppIDAndPubUserID returns a device of given appID and pubUserID.
func (r *Repository) GetByAppIDAndPubUserID(appID int64, pubUserID string) (*device.Device, error) {
	dbDevice, err := r.source.GetByAppIDAndPubUserID(appID, pubUserID)
	if dbDevice != nil {
		return r.mapper.dbDeviceToDevice(*dbDevice), nil
	}
	return nil, err
}

// UpsertDevice returns updated device after inserting or updating the device's members.
func (r *Repository) UpsertDevice(d device.Device) (*device.Device, error) {
	dbDevice, err := r.source.UpsertDevice(*r.mapper.deviceToDBDevice(d))
	if dbDevice != nil {
		return r.mapper.dbDeviceToDevice(*dbDevice), err
	}
	return nil, err
}

// New returns new device repository.
func New(source dbdevice.DBSource) *Repository {
	return &Repository{source, mapper{}}
}
