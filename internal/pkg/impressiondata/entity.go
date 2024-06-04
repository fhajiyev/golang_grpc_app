package impressiondata

// ImpressionData struct definition
type ImpressionData struct {
	IFA             string  `json:"i"`
	CampaignID      int64   `json:"c"`
	UnitID          int64   `json:"u"`
	DeviceID        int64   `json:"d"`
	UnitDeviceToken string  `json:"udt"`
	Country         string  `json:"cou"`
	Gender          *string `json:"sex,omitempty"`
	YearOfBirth     *int    `json:"yob,omitempty"`
}

const (
	impressionDataAESKey  = "4i3jfkdie0998754"
	impressionDataURLSafe = true
)
