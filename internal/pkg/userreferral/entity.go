package userreferral

// DeviceUser stores user with referral information
type DeviceUser struct {
	ID         int64
	DeviceID   int64
	Code       string
	ReferrerID int64
	IsVerified bool
}

// GiveReferralRewardRequestIngredients is request body for calling GiveReferralReward API
type GiveReferralRewardRequestIngredients struct {
	RefereeDeviceID  int64
	RefereeReward    int
	ReferrerDeviceID int64
	ReferrerReward   int
	JWT              string
	TitleForReferral TitleForReferral
}

// TitleForReferral is titles for referral
type TitleForReferral struct {
	TitleForReferee     string
	TitleForReferrer    string
	TitleForMaxReferrer string
}

// CreateReferralIngredients is struct for calling CreateReferral
type CreateReferralIngredients struct {
	DeviceID         int64
	AppID            int64
	UnitDeviceToken  string
	Code             string
	JWT              string
	VerifyURL        string
	RewardAmount     int
	MaxReferral      int
	TitleForReferral TitleForReferral
}
