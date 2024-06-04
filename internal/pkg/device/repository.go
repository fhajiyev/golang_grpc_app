package device

// ProfileRepository is the repository interface of device package
type ProfileRepository interface {
	GetByID(deviceID int64) (*Profile, error)
	Save(dp Profile) error
	SavePackage(dp Profile) error
	SaveUnitRegisteredSeconds(dp Profile) error
	Delete(dp Profile) error
}

// ActivityRepository interface definition
type ActivityRepository interface {
	GetByID(deviceID int64) (*Activity, error)
	Save(deviceID int64, campaignID int64, activityType ActivityType) error
}

// Repository interface definition
type Repository interface {
	GetByID(deviceID int64) (*Device, error)
	GetByAppIDAndIFA(appID int64, ifa string) (*Device, error)
	GetByAppIDAndPubUserID(appID int64, pubUserID string) (*Device, error)
	UpsertDevice(device Device) (du *Device, err error)
}
