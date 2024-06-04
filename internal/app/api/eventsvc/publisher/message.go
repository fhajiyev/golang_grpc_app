package publisher

type baseMessage struct {
	ResourceID   int64  `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	Event        string `json:"event"`
}

type eventMessage struct {
	baseMessage
	Extra eventExtra `json:"extra"`
}

type eventExtra struct {
	Reward       reward       `json:"reward"`
	ResourceData ResourceData `json:"resource"`
	UnitData     UnitData     `json:"unit"`
}

type reward struct {
	TransactionID string `json:"transaction_id"`
}

// ResourceData contains resource-related data
type ResourceData struct {
	ID             int64                  `json:"id"`
	Name           string                 `json:"name"`
	OrganizationID int64                  `json:"organization_id"`
	RevenueType    string                 `json:"revenue_type"`
	IsMedia        bool                   `json:"is_media"`
	Extra          map[string]interface{} `json:"extra"`
}

// UnitData contains unit-related data
type UnitData struct {
	ID int64 `json:"id"`
}
