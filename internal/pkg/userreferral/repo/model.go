package repo

import "time"

// DBDeviceUser stores user with referral information in DB
type DBDeviceUser struct {
	ID         int64 `gorm:"primary_key"`
	DeviceID   int64
	Code       string
	ReferrerID int64
	IsVerified bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName is table name in db for DBDeviceUser
func (DBDeviceUser) TableName() string {
	return "device_user"
}
