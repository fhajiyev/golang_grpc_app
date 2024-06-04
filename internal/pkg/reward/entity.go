package reward

// RequestIngredients is POST data
type RequestIngredients struct {
	AppID           int64
	UnitID          int64
	DeviceID        int64
	IFA             string
	UnitDeviceToken string

	CampaignID      int64
	CampaignType    string
	CampaignName    string
	CampaignOwnerID *string
	CampaignIsMedia int
	Slot            int

	Reward     int
	BaseReward int
	ClickType  ClickType
	Checksum   string
}

// Point entity
type Point struct {
	DeviceID   int64
	Version    int64
	CampaignID int64
	Type       string
	CreatedAt  int64
}
