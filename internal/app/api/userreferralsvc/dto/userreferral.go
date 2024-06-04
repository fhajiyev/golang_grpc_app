package dto

// DeviceUser struct for API response
type DeviceUser struct {
	ID         int64  `json:"id"`
	DeviceID   int64  `json:"device_id"`
	Code       string `json:"code"`
	ReferrerID int64  `json:"referrer_id"`
	IsVerified bool   `json:"is_verified"`
}

// RewardConfig struct for API response
type RewardConfig struct {
	AppID   int64 `json:"app_id"`
	Enabled bool  `json:"enabled"`
	Amount  int   `json:"amount"`
	Ended   bool  `json:"ended"`
}

// GetUserResponse is response for GetUser API
type GetUserResponse struct {
	User   *DeviceUser   `json:"user"`
	Config *RewardConfig `json:"referral_config"`
}
