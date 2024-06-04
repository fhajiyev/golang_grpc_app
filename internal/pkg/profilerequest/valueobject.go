package profilerequest

// Account defines pixel attributes and events passed from tracking SDK using GET.
type Account struct {
	AppID     int64
	IFA       string
	AccountID int64
	CookieID  string
	AppUserID string
}
