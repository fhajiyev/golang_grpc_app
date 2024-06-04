package event

import "time"

// Event struct
type Event struct {
	Type         string
	TrackingURLs []string
	Reward       *Reward
	Extra        map[string]string
}

// Events is array of Event
type Events []Event

// Type is type of event
const (
	TypeImpressed string = "impressed"
	TypeClicked   string = "clicked"
	TypeLanded    string = "landed"

	StatusReceivable      string = "receivable"
	StatusAlreadyReceived string = "already_received"
)

// Reward struct
type Reward struct {
	Amount         int64
	Status         string
	IssueMethod    string
	StatusCheckURL string
	TTL            int64
	Extra          map[string]string
}

// Token are derived from token
type Token struct {
	Resource      Resource
	EventType     string
	UnitID        int64
	TransactionID string
}

// TokenExpiration is used to initialize jwe.Manager in server
const TokenExpiration time.Duration = 3 * 24 * time.Hour

const (
	logTrackURLActivityMethodSave = "save"
	logTrackURLActivityMethodGet  = "get"
)
