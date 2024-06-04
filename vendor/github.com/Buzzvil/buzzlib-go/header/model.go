package header

// Auth struct contains identifiable data for a user
type Auth struct {
	AppID           int64
	AccountID       int64
	PublisherUserID string
	IFA             string
}

// BuzzUserAgent struct contains identifiable data for a user agent
type BuzzUserAgent struct {
	SDKName        string
	SDKVersion     string
	SDKVersionCode int64
	AppName        string
	AppVersion     string
	AppVersionCode int64
	OS             string
	OSVersion      string
	OSVersionCode  int64
	Model          string
	Manufacturer   string
	Device         string
	Brand          string
	Product        string
}
