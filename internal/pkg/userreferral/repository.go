package userreferral

// Repository interface definition
type Repository interface {
	CreateUser(deviceID int64, code string, isVerified bool) (*DeviceUser, error)
	UpdateUserReferrerID(userID int64, referrerID int64) error
	GetUserByDevice(deviceID int64) (*DeviceUser, error)
	GetUserByCode(code string) (*DeviceUser, error)
	GetReferralCountByUser(referrerID int64) (int, error)

	IsVerifiedDevice(verifyURL string, udt string) error
	GiveReferralReward(ingr GiveReferralRewardRequestIngredients) error
}
