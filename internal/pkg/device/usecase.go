package device

import "errors"

// UseCase interface definition
type UseCase interface {
	GetProfile(deviceID int64) (*Profile, error)
	GetActivity(deviceID int64) (*Activity, error)
	SaveActivity(deviceID int64, campaignID int64, activityType ActivityType) error
	SaveProfile(dp Profile) error
	SaveProfilePackage(dp Profile) error
	SaveProfileUnitRegisteredSeconds(dp Profile) error
	DeleteProfile(dp Profile) error
	GetByID(deviceID int64) (*Device, error)
	GetByParams(params Params) (*Device, error)
	UpsertDevice(device Device) (*Device, error)

	ValidateUnitDeviceToken(unitDeviceToken string) (bool, error)
}

type deviceUseCase struct {
	repo    Repository
	proRepo ProfileRepository
	actRepo ActivityRepository
}

// GetActivity func definition
func (u *deviceUseCase) GetActivity(deviceID int64) (*Activity, error) {
	return u.actRepo.GetByID(deviceID)
}

// SaveActivity func definition
func (u *deviceUseCase) SaveActivity(deviceID int64, campaignID int64, activityType ActivityType) error {
	return u.actRepo.Save(deviceID, campaignID, activityType)
}

// GetProfile func definition
func (u *deviceUseCase) GetProfile(deviceID int64) (*Profile, error) {
	return u.proRepo.GetByID(deviceID)
}

// SaveProfile func definition
func (u *deviceUseCase) SaveProfile(dp Profile) error {
	return u.proRepo.Save(dp)
}

// SaveProfilePackage func definition
func (u *deviceUseCase) SaveProfilePackage(dp Profile) error {
	return u.proRepo.SavePackage(dp)
}

// SaveProfileUnitRegisteredSecond func definition
func (u *deviceUseCase) SaveProfileUnitRegisteredSeconds(dp Profile) error {
	return u.proRepo.SaveUnitRegisteredSeconds(dp)
}

// DeleteProfile func definition
func (u *deviceUseCase) DeleteProfile(dp Profile) error {
	return u.proRepo.Delete(dp)
}

// GetByID func definition
func (u *deviceUseCase) GetByID(deviceID int64) (*Device, error) {
	if deviceID == 0 {
		return nil, InvalidArgumentError{ArgName: "deviceID", ArgValue: 0}
	}
	return u.repo.GetByID(deviceID)
}

// GetByParams func definition
func (u *deviceUseCase) GetByParams(params Params) (*Device, error) {
	if params.PubUserID != "" {
		if d, err := u.repo.GetByAppIDAndPubUserID(params.AppID, params.PubUserID); d != nil {
			return d, err
		}
	}
	return u.repo.GetByAppIDAndIFA(params.AppID, params.IFA)
}

// UpsertDevice func definition
func (u *deviceUseCase) UpsertDevice(device Device) (*Device, error) {
	return u.repo.UpsertDevice(device)
}

// ValidateUnitDeviceToken validates unit device token
func (u *deviceUseCase) ValidateUnitDeviceToken(unitDeviceToken string) (bool, error) {
	validators := []func(string) (bool, error){
		u.validateEmpty,
		u.validateStartWithNilCharacter,
	}

	for _, validator := range validators {
		ok, err := validator(unitDeviceToken)
		if !ok {
			return false, err
		}
	}

	return true, nil
}

func (u *deviceUseCase) validateEmpty(unitDeviceToken string) (bool, error) {
	if len(unitDeviceToken) == 0 {
		return false, errors.New("empty unit device token")
	}

	return true, nil
}

func (u *deviceUseCase) validateStartWithNilCharacter(unitDeviceToken string) (bool, error) {
	if unitDeviceToken[0] == 0 {
		return false, errors.New("unit device token starts with nil character")
	}
	return true, nil
}

var _ UseCase = &deviceUseCase{}

// NewUseCase returns new device usecase.
func NewUseCase(repo Repository, profileRepo ProfileRepository, activityRepo ActivityRepository) UseCase {
	return &deviceUseCase{repo, profileRepo, activityRepo}
}
