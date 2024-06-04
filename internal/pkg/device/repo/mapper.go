package repo

import (
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
)

type mapper struct {
}

func (m *mapper) dbDeviceToDevice(dbDevice dbdevice.Device) *device.Device {
	return &(device.Device{
		ID:              dbDevice.ID,
		AppID:           dbDevice.AppID,
		UnitDeviceToken: dbDevice.UnitDeviceToken,
		IFA:             dbDevice.IFA,
		Address:         dbDevice.Address,
		Birthday:        dbDevice.Birthday,
		Carrier:         dbDevice.Carrier,
		DeviceName:      dbDevice.DeviceName,
		Resolution:      dbDevice.Resolution,
		YearOfBirth:     dbDevice.YearOfBirth,
		SDKVersion:      dbDevice.SDKVersion,
		Sex:             dbDevice.Sex,
		Packages:        dbDevice.Packages,
		PackageName:     dbDevice.PackageName,
		SignupIP:        dbDevice.SignupIP,
		SerialNumber:    dbDevice.SerialNumber,
		CreatedAt:       dbDevice.CreatedAt,
		UpdatedAt:       dbDevice.UpdatedAt,
	})
}

func (m *mapper) deviceToDBDevice(device device.Device) *dbdevice.Device {
	return &(dbdevice.Device{
		ID:              device.ID,
		AppID:           device.AppID,
		UnitDeviceToken: device.UnitDeviceToken,
		IFA:             device.IFA,
		Address:         device.Address,
		Birthday:        device.Birthday,
		Carrier:         device.Carrier,
		DeviceName:      device.DeviceName,
		Resolution:      device.Resolution,
		YearOfBirth:     device.YearOfBirth,
		SDKVersion:      device.SDKVersion,
		Sex:             device.Sex,
		Packages:        device.Packages,
		PackageName:     device.PackageName,
		SignupIP:        device.SignupIP,
		SerialNumber:    device.SerialNumber,
		CreatedAt:       device.CreatedAt,
		UpdatedAt:       device.UpdatedAt,
	})
}
