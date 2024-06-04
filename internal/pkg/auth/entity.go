package auth

// Auth contains authenticated information.
type Auth struct {
	AccountID       int64
	AppID           int64
	PublisherUserID string
	IFA             string
}

// Identifier contains account information.
type Identifier struct {
	AppID           int64
	PublisherUserID string
	IFA             string
}
