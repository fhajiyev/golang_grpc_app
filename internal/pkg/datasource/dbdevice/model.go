package dbdevice

import (
	"time"
)

// Device type definition
type Device struct {
	ID              int64  //`gorm:"primary_key"`
	AppID           int64  `gorm:"unique_index:index_udt,index_ifa"`
	UnitDeviceToken string `gorm:"type:varchar(255);unique_index:index_udt"`
	IFA             string `gorm:"type:varchar(45);unique_index:index_ifa"`
	Address         *string
	Birthday        *time.Time
	Carrier         *string
	DeviceName      string
	Resolution      string
	YearOfBirth     *int
	SDKVersion      *int
	Sex             *string
	Packages        *string
	PackageName     *string
	SignupIP        int64
	SerialNumber    *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TableName func definition
func (Device) TableName() string {
	return "device"
}

/*
	기존 레코드 d에, 요청으로 부터 온 new device를 머지하는 작업
	new device에 데이터가 있으면 latest data로 간주하고 record에 overwrite,
	데이터가 없으면 기존 record에서 가진 데이터로 유지함
	DeviceID는 언제나 new device에서 0이고, AppID는 언제나 같으므로 update를 시도하지않음

	UnitDeviceToken, IFA, PackageName은 변경시 DeviceUpdateHistory를 생성함
*/
func (d *Device) updateWithHistory(new Device) []DeviceUpdateHistory {
	if new.Address != nil && *new.Address != "" {
		d.Address = new.Address
	}
	if new.Birthday != nil {
		d.Birthday = new.Birthday
	}
	if new.Carrier != nil && *new.Carrier != "" {
		d.Carrier = new.Carrier
	}
	if new.DeviceName != "" {
		d.DeviceName = new.DeviceName
	}
	if new.Resolution != "" {
		d.Resolution = new.Resolution
	}
	if new.YearOfBirth != nil && *new.YearOfBirth != 0 {
		d.YearOfBirth = new.YearOfBirth
	}
	if new.SDKVersion != nil && *new.SDKVersion != 0 {
		d.SDKVersion = new.SDKVersion
	}
	if new.Sex != nil && *new.Sex != "" {
		d.Sex = new.Sex
	}
	if new.Packages != nil && *new.Packages != "" {
		d.Packages = new.Packages
	}
	if new.SignupIP != 0 {
		d.SignupIP = new.SignupIP
	}
	if new.SerialNumber != nil && *new.SerialNumber != "" {
		d.SerialNumber = new.SerialNumber
	}

	var history *DeviceUpdateHistory
	var histories []DeviceUpdateHistory
	d.IFA, history = d.updateFieldWithHistory(&new.IFA, &d.IFA, historyFieldNameIFA)
	if history != nil {
		histories = append(histories, *history)
	}
	d.UnitDeviceToken, history = d.updateFieldWithHistory(&new.UnitDeviceToken, &d.UnitDeviceToken, historyFieldNameUnitDeviceToken)
	if history != nil {
		histories = append(histories, *history)
	}
	packageName, history := d.updateFieldWithHistory(new.PackageName, d.PackageName, historyFieldNamePackageName)
	d.PackageName = &packageName
	if history != nil {
		histories = append(histories, *history)
	}

	return histories
}

func (d *Device) updateFieldWithHistory(requestValue *string, originalValue *string, fieldName string) (string, *DeviceUpdateHistory) {
	isStringEmpty := func(s *string) bool {
		return s == nil || *s == ""
	}
	getString := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	RequestValueString := getString(requestValue)
	OriginalValueString := getString(originalValue)

	if isStringEmpty(requestValue) {
		return OriginalValueString, nil
	} else if isStringEmpty(originalValue) { // db에 값이 비어있으면 requestValue를 저장, history업데이트 X
		return RequestValueString, nil
	} else if RequestValueString == OriginalValueString {
		return RequestValueString, nil
	}

	history := newDeviceHistory(d.ID, fieldName, OriginalValueString, RequestValueString)
	return RequestValueString, &history // 값이 다른경우 업데이트기록 생성
}

// DeviceUpdateHistory type definition
type DeviceUpdateHistory struct {
	ID           int64 `gorm:"primary_key"`
	DeviceID     int64
	UpdatedField string
	FromValue    string
	ToValue      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName func definition
func (DeviceUpdateHistory) TableName() string {
	return "device_update_history"
}

const (
	historyFieldNameIFA             = "ifa"
	historyFieldNameUnitDeviceToken = "unit_device_token"
	historyFieldNamePackageName     = "package_name"
)

func newDeviceHistory(deviceID int64, fieldName string, from string, to string) DeviceUpdateHistory {
	return DeviceUpdateHistory{
		DeviceID:     deviceID,
		UpdatedField: fieldName,
		FromValue:    from,
		ToValue:      to,
	}
}
