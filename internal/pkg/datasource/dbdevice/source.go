package dbdevice

// DBSource interface definition
type DBSource interface {
	GetByID(deviceID int64) (*Device, error)
	GetByAppIDAndPubUserID(appID int64, pubUserID string) (*Device, error)
	GetByAppIDAndIFA(appID int64, ifa string) (*Device, error)
	UpsertDevice(device Device) (*Device, error)
}
