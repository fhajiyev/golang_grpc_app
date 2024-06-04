package repo

type adDetail struct {
	ID             int64                  `json:"id"`
	ItemName       string                 `json:"item_name"`
	OrganizationID int64                  `json:"organization_id"`
	RevenueType    string                 `json:"revenue_type"`
	ExtraData      map[string]interface{} `json:"extra_data"`
}
